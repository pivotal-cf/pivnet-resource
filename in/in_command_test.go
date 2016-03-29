package in_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("In", func() {
	var (
		server        *ghttp.Server
		file1Contents string

		downloadDir string

		ginkgoLogger logger.Logger

		productVersion  string
		etag            string
		versionWithETag string

		inRequest              concourse.InRequest
		inCommand              *in.InCommand
		pivnetReleasesResponse *pivnet.ReleasesResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		productVersion = "C"
		etag = "etag-0"

		var err error
		versionWithETag, err = versions.CombineVersionAndETag(productVersion, etag)
		Expect(err).NotTo(HaveOccurred())

		releaseID := 1234
		file1URLPath := "/file1"
		file1URL := fmt.Sprintf("%s%s", server.URL(), file1URLPath)
		file1Contents = ""
		statusCode := http.StatusOK

		pivnetReleasesResponse = &pivnet.ReleasesResponse{
			Releases: []pivnet.Release{
				{
					Version: "A",
				},
				{
					Version: productVersion,
					ID:      releaseID,
					Links: &pivnet.Links{
						ProductFiles: map[string]string{
							"href": file1URL,
						},
					},
				},
				{
					Version: "B",
				},
			},
		}

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, pivnetReleasesResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf(
						"%s/products/%s/releases/%d/eula_acceptance",
						apiPrefix,
						productSlug,
						releaseID,
					),
				),
				ghttp.RespondWith(http.StatusOK, ""),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					file1URLPath,
				),
				ghttp.RespondWith(http.StatusOK, file1Contents),
			),
		)

		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				Endpoint:    server.URL(),
			},
			Version: concourse.Version{
				versionWithETag,
			},
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2"
		inCommand = in.NewInCommand(binaryVersion, ginkgoLogger, downloadDir)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the version comes from concourse", func() {
		It("creates a version file with the downloaded version and etag", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			versionFilepath := filepath.Join(downloadDir, "version")
			versionContents, err := ioutil.ReadFile(versionFilepath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(versionContents)).To(Equal(versionWithETag))
		})

		It("does not download any of the files in the specified release", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			files, err := ioutil.ReadDir(downloadDir)
			Expect(err).ShouldNot(HaveOccurred())

			// the version file will always exist
			Expect(len(files)).To(Equal(1))
			Expect(files[0].Name()).To(Equal("version"))
		})
	})

	Context("when the version is specified by the user", func() {
		BeforeEach(func() {
			inRequest.Source.ProductVersion = "1.2.5"
			pivnetReleasesResponse.Releases[1].Version = "1.2.5"
		})

		It("requests the configured release", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			versionFilepath := filepath.Join(downloadDir, "version")
			versionContents, err := ioutil.ReadFile(versionFilepath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(versionContents)).To(Equal("1.2.5"))
		})
	})

	Context("when no api token is provided", func() {
		BeforeEach(func() {
			inRequest.Source.APIToken = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
		})
	})

	Context("when no product slug is provided", func() {
		BeforeEach(func() {
			inRequest.Source.ProductSlug = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
		})
	})

	Context("when no product version is provided", func() {
		BeforeEach(func() {
			inRequest.Source.ProductVersion = ""
			inRequest.Version.ProductVersion = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*product_version.*provided"))
		})
	})

	Context("when version is provided without etag", func() {
		BeforeEach(func() {
			inRequest.Version = concourse.Version{
				ProductVersion: productVersion,
			}
		})

		It("returns without error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
