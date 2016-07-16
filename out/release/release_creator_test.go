package release_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCreator", func() {
	var (
		fetcherClient       *releasefakes.Fetcher
		pivnetClient        *releasefakes.ReleaseClient
		fakeSemverConverter *releasefakes.FakeSemverConverter
		logging             *log.Logger

		creator release.ReleaseCreator

		sortBy                      concourse.SortBy
		fetchReturnedProductVersion string
		existingProductVersion      string
		eulaSlug                    string
		releaseType                 string
	)

	BeforeEach(func() {
		pivnetClient = &releasefakes.ReleaseClient{}
		fetcherClient = &releasefakes.Fetcher{}
		logging = log.New(ioutil.Discard, "it doesn't matter", 0)
		fakeSemverConverter = &releasefakes.FakeSemverConverter{}

		sortBy = concourse.SortByNone

		existingProductVersion = "existing-product-version"
		fetchReturnedProductVersion = "a-product-version"
		eulaSlug = "magic-slug"
		releaseType = "some-release-type"

		pivnetClient.EULAsReturns([]pivnet.EULA{{Slug: eulaSlug}}, nil)
		pivnetClient.ReleaseTypesReturns([]string{releaseType}, nil)
		pivnetClient.ReleasesForProductSlugReturns([]pivnet.Release{{Version: existingProductVersion}}, nil)
		pivnetClient.CreateReleaseReturns(pivnet.Release{ID: 1337}, nil)
	})

	Describe("Create", func() {
		JustBeforeEach(func() {
			meta := metadata.Metadata{
				Release: &metadata.Release{
					Controlled: true,
				},
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

			source := concourse.Source{
				SortBy: sortBy,
			}

			creator = release.NewReleaseCreator(
				pivnetClient,
				fetcherClient,
				fakeSemverConverter,
				logging,
				meta,
				false,
				params,
				source,
				"/some/sources/dir",
				"some-product-slug",
			)

			fetcherClient.FetchStub = func(name string, sourcesDir string, file string) string {
				switch name {
				case "EULASlug":
					return eulaSlug
				case "ReleaseType":
					return releaseType
				case "Version":
					return fetchReturnedProductVersion
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
		})

		Context("when the release does not exist", func() {
			BeforeEach(func() {
				pivnetClient.ProductVersionsReturns([]string{"a version that has not been uploaded"}, nil)
			})

			It("constructs the release", func() {
				r, err := creator.Create()
				Expect(err).NotTo(HaveOccurred())

				Expect(r).To(Equal(pivnet.Release{ID: 1337}))

				Expect(pivnetClient.EULAsCallCount()).To(Equal(1))

				key, sourcesDir, fileName := fetcherClient.FetchArgsForCall(0)
				Expect(key).To(Equal("Version"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-version-file"))

				key, sourcesDir, fileName = fetcherClient.FetchArgsForCall(1)
				Expect(key).To(Equal("EULASlug"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-eula-slug-file"))

				Expect(pivnetClient.ReleaseTypesCallCount()).To(Equal(1))

				key, sourcesDir, fileName = fetcherClient.FetchArgsForCall(2)
				Expect(key).To(Equal("ReleaseType"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(fileName).To(Equal("some-release-type-file"))

				Expect(pivnetClient.ReleasesForProductSlugArgsForCall(0)).To(Equal("some-product-slug"))

				slug, releases := pivnetClient.ProductVersionsArgsForCall(0)
				Expect(slug).To(Equal("some-product-slug"))
				Expect(releases).To(Equal([]pivnet.Release{{Version: existingProductVersion}}))

				Expect(pivnetClient.CreateReleaseArgsForCall(0)).To(Equal(pivnet.CreateReleaseConfig{
					ProductSlug:     "some-product-slug",
					ReleaseType:     releaseType,
					EULASlug:        eulaSlug,
					ProductVersion:  fetchReturnedProductVersion,
					Description:     "wow, a description",
					ReleaseNotesURL: "some-url",
					ReleaseDate:     "1/17/2016",
					Controlled:      true,
				}))
			})

			Context("when an error occurs", func() {
				Context("when pivnet fails getting releases for a product slug", func() {
					BeforeEach(func() {
						pivnetClient.ReleasesForProductSlugReturns([]pivnet.Release{}, errors.New("product slug error"))
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("product slug error")))
					})
				})

				Context("when pivnet fails getting product versions", func() {
					BeforeEach(func() {
						pivnetClient.ProductVersionsReturns([]string{""}, errors.New("missing product version"))
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("missing product version")))
					})
				})

				Context("when pivnet fails fetching eulas", func() {
					BeforeEach(func() {
						pivnetClient.EULAsReturns([]pivnet.EULA{}, errors.New("failed getting eulas"))
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("failed getting eulas")))
					})
				})

				Context("when the metadata does not contain the eula slug", func() {
					BeforeEach(func() {
						pivnetClient.EULAsReturns([]pivnet.EULA{{Slug: "a-failing-slug"}}, nil)
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("provided eula_slug: 'magic-slug' must be one of: ['a-failing-slug']")))
					})
				})

				Context("when pivnet fails fetching release types", func() {
					BeforeEach(func() {
						pivnetClient.ReleaseTypesReturns([]string{""}, errors.New("failed fetching release types"))
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("failed fetching release types")))
					})
				})

				Context("when the metadata does not contain the release type", func() {
					BeforeEach(func() {
						pivnetClient.ReleaseTypesReturns([]string{"a-missing-release-type"}, nil)
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("provided release_type: 'some-release-type' must be one of: ['a-missing-release-type']")))
					})
				})

				Context("when the release cannot be created", func() {
					BeforeEach(func() {
						pivnetClient.CreateReleaseReturns(pivnet.Release{}, errors.New("cannot create release"))
					})

					It("returns an error", func() {
						_, err := creator.Create()
						Expect(err).To(MatchError(errors.New("cannot create release")))
					})
				})
			})
		})

		Context("when the release already exists", func() {
			BeforeEach(func() {
				fetcherClient.FetchReturns(fetchReturnedProductVersion)
				pivnetClient.ReleasesForProductSlugReturns([]pivnet.Release{{Version: fetchReturnedProductVersion}}, nil)
				pivnetClient.ProductVersionsReturns([]string{fetchReturnedProductVersion}, nil)
			})

			It("returns a error", func() {
				_, err := creator.Create()
				Expect(err).To(MatchError(fmt.Errorf("release already exists with version: %s", fetchReturnedProductVersion)))
			})
		})

		Context("when sorting by semver", func() {
			BeforeEach(func() {
				sortBy = concourse.SortBySemver
			})

			Context("when release is not valid semver", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = fmt.Errorf("semver parse error")
					fakeSemverConverter.ToValidSemverReturns(semver.Version{}, expectedErr)
				})

				It("returns an error", func() {
					_, err := creator.Create()
					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
