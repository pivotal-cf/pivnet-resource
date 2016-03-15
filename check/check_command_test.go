package check_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/check"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

var _ = Describe("Check", func() {
	var (
		server         *ghttp.Server
		pivnetResponse string

		tempDir     string
		logFilePath string

		ginkgoLogger logger.Logger

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		pivnetResponse =
			`{"releases": [{"id":1,"version": "A"},{"id":2,"version":"C"},{"id":3,"version":"B"}]}`

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
				ghttp.RespondWith(http.StatusOK, pivnetResponse),
			),
		)

		for i := 0; i < 3; i++ {
			etag := fmt.Sprintf(`"etag-%d"`, i)
			etagHeader := http.Header{"ETag": []string{etag}}
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"HEAD",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, i+1)),
					ghttp.RespondWith(http.StatusOK, pivnetResponse, etagHeader),
				),
			)
		}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "pivnet-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		binaryVersion := "v0.1.2-unit-tests"

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				Endpoint:    server.URL(),
			},
		}

		sanitized := concourse.SanitizedSource(checkRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		checkCommand = check.NewCheckCommand(binaryVersion, ginkgoLogger, logFilePath)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns the most recent version without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(HaveLen(1))
		Expect(response[0].ProductVersion).To(Equal("A#etag-0"))
	})

	Context("when no api token is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.APIToken = ""
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
		})
	})

	Context("when no product slug is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.ProductSlug = ""
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
		})
	})

	Context("when no releases are returned", func() {
		BeforeEach(func() {
			pivnetResponse = fmt.Sprintf(`{"releases": []}`)

			server.Reset()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWith(http.StatusOK, pivnetResponse),
				),
			)
		})

		It("returns empty response without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(BeEmpty())
		})
	})

	Context("when log files already exist", func() {
		var (
			otherFilePath1 string
			otherFilePath2 string
		)

		BeforeEach(func() {
			otherFilePath1 = filepath.Join(tempDir, "pivnet-resource-check.log1")
			err := ioutil.WriteFile(otherFilePath1, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			otherFilePath2 = filepath.Join(tempDir, "pivnet-resource-check.log2")
			err = ioutil.WriteFile(otherFilePath2, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes the other log files", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(otherFilePath1)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(otherFilePath2)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(logFilePath)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when there is an error getting product versions", func() {
		BeforeEach(func() {
			server.Reset()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusNotFound, ""),
				),
			)
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when a version is provided", func() {
		BeforeEach(func() {
			checkRequest.Version = concourse.Version{
				"B#etag-2",
			}
		})

		It("returns the most recent version", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(2))
			Expect(response[0].ProductVersion).To(Equal("C#etag-1"))
			Expect(response[1].ProductVersion).To(Equal("A#etag-0"))
		})
	})
})
