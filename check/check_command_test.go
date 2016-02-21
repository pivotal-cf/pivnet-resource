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

		tempDir string
		logFile *os.File

		version      string
		ginkgoLogger logger.Logger

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		pivnetResponse = fmt.Sprintf(`{"releases": [{"version": "1234"}]}`)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
				ghttp.RespondWith(http.StatusOK, pivnetResponse),
			),
		)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath := filepath.Join(tempDir, "check.log")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		logFile, err = os.Open(logFilePath)
		Expect(err).NotTo(HaveOccurred())

		version = "some-version"

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

		checkCommand = check.NewCheckCommand(version, ginkgoLogger, logFile)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("runs without error", func() {
		_, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())
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
})
