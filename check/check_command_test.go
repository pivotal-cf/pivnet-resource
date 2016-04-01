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
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("Check", func() {
	var (
		server         *ghttp.Server
		pivnetResponse pivnet.ReleasesResponse

		tempDir     string
		logFilePath string

		ginkgoLogger logger.Logger

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand

		releaseTypes []string

		releaseTypesResponse       pivnet.ReleaseTypesResponse
		releaseTypesResponseStatus int

		allReleases      []pivnet.Release
		filteredReleases []pivnet.Release

		releasesResponseStatus int
		etagResponseStatus     int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		releaseTypes = []string{
			"foo release",
			"bar",
			"third release type",
		}

		releaseTypesResponse = pivnet.ReleaseTypesResponse{
			ReleaseTypes: releaseTypes,
		}
		releaseTypesResponseStatus = http.StatusOK

		allReleases = []pivnet.Release{
			{
				ID:          1,
				Version:     "A",
				ReleaseType: releaseTypes[0],
			},
			{
				ID:          2,
				Version:     "C",
				ReleaseType: releaseTypes[1],
			},
			{
				ID:          3,
				Version:     "B",
				ReleaseType: releaseTypes[2],
			},
		}
		releasesResponseStatus = http.StatusOK
		etagResponseStatus = http.StatusOK

		pivnetResponse = pivnet.ReleasesResponse{
			Releases: allReleases,
		}

		filteredReleases = allReleases

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				Endpoint:    server.URL(),
			},
		}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "pivnet-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/releases/release_types", apiPrefix)),
				ghttp.RespondWithJSONEncoded(releaseTypesResponseStatus, releaseTypesResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
				ghttp.RespondWithJSONEncoded(releasesResponseStatus, pivnetResponse),
			),
		)

		for _, release := range filteredReleases {
			etag := fmt.Sprintf(`"etag-%d"`, release.ID)
			etagHeader := http.Header{"ETag": []string{etag}}
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID),
					),
					ghttp.RespondWith(etagResponseStatus, nil, etagHeader),
				),
			)
		}

		binaryVersion := "v0.1.2-unit-tests"

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

		expectedVersionWithEtag, err := versions.CombineVersionAndETag(
			allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(HaveLen(1))
		Expect(response[0].ProductVersion).To(Equal(expectedVersionWithEtag))
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
			pivnetResponse = pivnet.ReleasesResponse{Releases: []pivnet.Release{}}
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

	Context("when there is an error getting release types", func() {
		BeforeEach(func() {
			releaseTypesResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when there is an error getting releases", func() {
		BeforeEach(func() {
			releasesResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when there is an error getting etag", func() {
		BeforeEach(func() {
			etagResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when a version is provided", func() {
		BeforeEach(func() {
			versionWithETag, err := versions.CombineVersionAndETag(
				allReleases[2].Version, fmt.Sprintf("etag-%d", allReleases[2].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			checkRequest.Version = concourse.Version{
				versionWithETag,
			}
		})

		It("returns the most recent version", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagC, err := versions.CombineVersionAndETag(
				allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagA, err := versions.CombineVersionAndETag(
				allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(2))
			Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
			Expect(response[1].ProductVersion).To(Equal(versionWithETagA))
		})
	})

	Context("when the release type is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = releaseTypes[1]

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		It("returns the most recent version with that release type", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagC, err := versions.CombineVersionAndETag(
				allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
		})

		Context("when the release type is invalid", func() {
			BeforeEach(func() {
				checkRequest.Source.ReleaseType = "not a valid release type"
			})

			It("returns an error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*release_type.*one of"))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[0]))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[1]))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[2]))
			})
		})
	})

	Context("when the product version is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = releaseTypes[1]

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		BeforeEach(func() {
			checkRequest.Source.ProductVersion = "C"
		})

		It("returns the newest release with that version without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagC, err := versions.CombineVersionAndETag(
				allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
		})
	})
})
