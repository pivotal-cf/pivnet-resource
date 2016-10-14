package release_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/blang/semver"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCreator", func() {
	var (
		pivnetClient        *releasefakes.ReleaseClient
		fakeSemverConverter *releasefakes.FakeSemverConverter
		logging             *log.Logger

		creator release.ReleaseCreator

		sourceReleaseType string
		sourceVersion     string
		sortBy            concourse.SortBy
		releaseVersion    string
		existingReleases  []pivnet.Release
		eulaSlug          string
		releaseType       pivnet.ReleaseType
	)

	BeforeEach(func() {
		pivnetClient = &releasefakes.ReleaseClient{}
		logging = log.New(ioutil.Discard, "it doesn't matter", 0)
		fakeSemverConverter = &releasefakes.FakeSemverConverter{}

		sortBy = concourse.SortByNone

		existingReleases = []pivnet.Release{
			{
				ID:      1234,
				Version: "1.8.1",
			},
		}

		releaseVersion = "1.8.3"
		eulaSlug = "magic-slug"
		releaseType = "some-release-type"

		sourceReleaseType = string(releaseType)
		sourceVersion = `1\.8\..*`

		pivnetClient.EULAsReturns([]pivnet.EULA{{Slug: eulaSlug}}, nil)
		pivnetClient.ReleaseTypesReturns([]pivnet.ReleaseType{releaseType}, nil)
		pivnetClient.ReleasesForProductSlugReturns(existingReleases, nil)
		pivnetClient.CreateReleaseReturns(pivnet.Release{ID: 1337}, nil)
	})

	Describe("Create", func() {
		JustBeforeEach(func() {
			meta := metadata.Metadata{
				Release: &metadata.Release{
					Controlled:      true,
					EULASlug:        eulaSlug,
					ReleaseType:     string(releaseType),
					Version:         releaseVersion,
					Description:     "wow, a description",
					ReleaseNotesURL: "some-url",
					ReleaseDate:     "1/17/2016",
				},
				ProductFiles: []metadata.ProductFile{
					{
						File:        "some/file",
						Description: "a description",
						UploadAs:    "a file",
					},
				},
			}

			params := concourse.OutParams{}

			source := concourse.Source{
				ReleaseType:    sourceReleaseType,
				ProductVersion: sourceVersion,
				SortBy:         sortBy,
			}

			creator = release.NewReleaseCreator(
				pivnetClient,
				fakeSemverConverter,
				logging,
				meta,
				params,
				source,
				"/some/sources/dir",
				"some-product-slug",
			)
		})

		It("constructs the release", func() {
			r, err := creator.Create()
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal(pivnet.Release{ID: 1337}))

			Expect(pivnetClient.EULAsCallCount()).To(Equal(1))

			Expect(pivnetClient.ReleasesForProductSlugArgsForCall(0)).To(Equal("some-product-slug"))

			Expect(pivnetClient.CreateReleaseArgsForCall(0)).To(Equal(pivnet.CreateReleaseConfig{
				ProductSlug:     "some-product-slug",
				ReleaseType:     string(releaseType),
				EULASlug:        eulaSlug,
				Version:         releaseVersion,
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
					Expect(err).To(MatchError(errors.New("provided EULA slug: 'magic-slug' must be one of: ['a-failing-slug']")))
				})
			})

			Context("when pivnet fails fetching release types", func() {
				BeforeEach(func() {
					pivnetClient.ReleaseTypesReturns([]pivnet.ReleaseType{}, errors.New("failed fetching release types"))
				})

				It("returns an error", func() {
					_, err := creator.Create()
					Expect(err).To(MatchError(errors.New("failed fetching release types")))
				})
			})

			Context("when the metadata does not contain the release type", func() {
				BeforeEach(func() {
					pivnetClient.ReleaseTypesReturns([]pivnet.ReleaseType{pivnet.ReleaseType("a-missing-release-type")}, nil)
				})

				It("returns an error", func() {
					_, err := creator.Create()
					Expect(err).To(MatchError(errors.New("provided release type: 'some-release-type' must be one of: ['a-missing-release-type']")))
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

		Context("when the release already exists", func() {
			BeforeEach(func() {
				releaseVersion = existingReleases[0].Version
			})

			It("returns a error", func() {
				_, err := creator.Create()
				Expect(err).To(MatchError(fmt.Errorf("release already exists with version: '%s'", releaseVersion)))
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

		Context("when release type does not match source config", func() {
			BeforeEach(func() {
				sourceReleaseType = "different release type"
				pivnetClient.ReleaseTypesReturns(
					[]pivnet.ReleaseType{releaseType, pivnet.ReleaseType(sourceReleaseType)},
					nil,
				)
			})

			It("returns an error", func() {
				_, err := creator.Create()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when source regex is invalid", func() {
			BeforeEach(func() {
				sourceVersion = `1\.[`
			})

			It("returns an error", func() {
				_, err := creator.Create()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when release version does not match source regex", func() {
			BeforeEach(func() {
				sourceVersion = `1\.7\..*`
			})

			It("returns an error", func() {
				_, err := creator.Create()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
