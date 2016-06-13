package release_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release/releasefakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCreator", func() {
	var (
		fetcherClient *releasefakes.Fetcher
		pivnetClient  *releasefakes.ReleaseClient
		logging       *releasefakes.Logging
		creator       release.ReleaseCreator
	)

	Describe("Create", func() {
		BeforeEach(func() {
			pivnetClient = &releasefakes.ReleaseClient{}
			fetcherClient = &releasefakes.Fetcher{}
			logging = &releasefakes.Logging{}

			meta := metadata.Metadata{
				Release: nil,
				ProductFiles: []metadata.ProductFile{
					{
						File:        "some/file",
						Description: "a description",
						UploadAs:    "a file",
					},
				},
			}

			params := concourse.OutParams{
				EULASlugFile:        "some-eula-slug-file",
				ReleaseTypeFile:     "some-release-type-file",
				VersionFile:         "some-version-file",
				ReleaseNotesURLFile: "some-release-notes-url-file",
				ReleaseDateFile:     "some-release-date-file",
			}

			creator = release.NewReleaseCreator(
				pivnetClient,
				fetcherClient,
				logging,
				meta,
				false,
				params,
				"/some/sources/dir",
				"some-product-slug",
			)
		})

		Context("when the release does not exist", func() {
			BeforeEach(func() {
				pivnetClient.EULAsReturns([]pivnet.EULA{{Slug: "magic-slug"}}, nil)
				fetcherClient.FetchStub = func(name string, sourcesDir string, file string) string {
					switch name {
					case "EULASlug":
						return "magic-slug"
					case "ReleaseType":
						return "some-release-type"
					case "Version":
						return "a-product-version"
					case "Description":
						return "wow, a description"
					case "ReleaseNotesURL":
						return "some-url"
					case "ReleaseDate":
						return "1/17/2016"
					default:
						panic("unexpected call")
					}
				}
				pivnetClient.ReleaseTypesReturns([]string{"some-release-type"}, nil)
				pivnetClient.ReleasesForProductSlugReturns([]pivnet.Release{{Version: "a version"}}, nil)
				pivnetClient.ProductVersionsReturns([]string{"a version that has not been uploaded"}, nil)
				pivnetClient.CreateReleaseReturns(pivnet.Release{ID: 1337}, nil)
			})

			It("constructs the release", func() {
				r, err := creator.Create()
				Expect(err).NotTo(HaveOccurred())

				Expect(r).To(Equal(pivnet.Release{ID: 1337}))

				message, _ := logging.DebugfArgsForCall(0)
				Expect(message).To(Equal("Getting all valid eulas\n"))

				Expect(pivnetClient.EULAsCallCount()).To(Equal(1))

				key, sourcesDir, fileName := fetcherClient.FetchArgsForCall(0)
				Expect(key).To(Equal("Version"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-version-file"))

				key, sourcesDir, fileName = fetcherClient.FetchArgsForCall(1)
				Expect(key).To(Equal("EULASlug"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-eula-slug-file"))

				message, _ = logging.DebugfArgsForCall(2)
				Expect(message).To(Equal("Getting all valid release types\n"))

				Expect(pivnetClient.ReleaseTypesCallCount()).To(Equal(1))

				key, sourcesDir, fileName = fetcherClient.FetchArgsForCall(2)
				Expect(key).To(Equal("ReleaseType"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-release-type-file"))

				Expect(pivnetClient.ReleasesForProductSlugArgsForCall(0)).To(Equal("some-product-slug"))

				slug, releases := pivnetClient.ProductVersionsArgsForCall(0)
				Expect(slug).To(Equal("some-product-slug"))
				Expect(releases).To(Equal([]pivnet.Release{{Version: "a version"}}))

				Expect(pivnetClient.CreateReleaseArgsForCall(0)).To(Equal(pivnet.CreateReleaseConfig{
					ProductSlug:     "some-product-slug",
					ReleaseType:     "some-release-type",
					EULASlug:        "magic-slug",
					ProductVersion:  "a-product-version",
					Description:     "wow, a description",
					ReleaseNotesURL: "some-url",
					ReleaseDate:     "1/17/2016",
				}))
			})
		})

		Context("when the release already exists", func() {
			BeforeEach(func() {
				fetcherClient.FetchReturns("an-existing-version")
				pivnetClient.ReleasesForProductSlugReturns([]pivnet.Release{{Version: "an-existing-version"}}, nil)
				pivnetClient.ProductVersionsReturns([]string{"an-existing-version"}, nil)
			})

			It("returns a error", func() {
				_, err := creator.Create()
				Expect(err).To(MatchError(errors.New("release already exists with version: an-existing-version")))
			})
		})
	})
})
