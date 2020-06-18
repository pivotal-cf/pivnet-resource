package release_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/blang/semver"
	"github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logger"
	"github.com/pivotal-cf/go-pivnet/v5/logshim"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCreator", func() {
	var (
		fakeLogger logger.Logger

		pivnetClient        *releasefakes.ReleaseClient
		fakeSemverConverter *releasefakes.FakeSemverConverter

		creator release.ReleaseCreator

		sourceReleaseType string
		sourceVersion     string
		sortBy            concourse.SortBy
		copyMetadata      bool
		releaseVersion    string
		existingReleases  []pivnet.Release
		eulaSlug          string
		productSlug       string
		releaseType       pivnet.ReleaseType
		params            concourse.OutParams
	)

	BeforeEach(func() {
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		pivnetClient = &releasefakes.ReleaseClient{}
		fakeSemverConverter = &releasefakes.FakeSemverConverter{}

		sortBy = concourse.SortByNone

		existingReleases = []pivnet.Release{
			{
				ID:      1234,
				Version: "1.8.1",
			},
		}

		productSlug = "some-product-slug"
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
		BeforeEach(func() {
			params = concourse.OutParams{}
		})

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

			source := concourse.Source{
				ReleaseType:    sourceReleaseType,
				ProductVersion: sourceVersion,
				SortBy:         sortBy,
				CopyMetadata:   copyMetadata,
			}

			creator = release.NewReleaseCreator(
				pivnetClient,
				fakeSemverConverter,
				fakeLogger,
				meta,
				params,
				source,
				"/some/sources/dir",
				productSlug,
			)
		})

		It("constructs the release", func() {
			r, err := creator.Create()
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal(pivnet.Release{ID: 1337}))

			Expect(pivnetClient.EULAsCallCount()).To(Equal(1))

			Expect(pivnetClient.ReleasesForProductSlugArgsForCall(0)).To(Equal(productSlug))

			Expect(pivnetClient.CreateReleaseArgsForCall(0)).To(Equal(pivnet.CreateReleaseConfig{
				ProductSlug:     productSlug,
				ReleaseType:     string(releaseType),
				EULASlug:        eulaSlug,
				Version:         releaseVersion,
				Description:     "wow, a description",
				ReleaseNotesURL: "some-url",
				ReleaseDate:     "1/17/2016",
				Controlled:      true,
				CopyMetadata:    copyMetadata,
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

			Context("when the Override parameter is set", func() {
				BeforeEach(func() {
					params.Override = true
				})

				It("deletes the release", func() {
					_, err := creator.Create()
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.DeleteReleaseCallCount()).To(Equal(1))

					invokedProductSlug, invokedRelease := pivnetClient.DeleteReleaseArgsForCall(0)
					Expect(invokedProductSlug).To(Equal(productSlug))
					Expect(invokedRelease).To(Equal(existingReleases[0]))
				})

				Context("when deleting the release returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = errors.New("some error")

						pivnetClient.DeleteReleaseReturns(expectedErr)
					})

					It("returns the error", func() {
						_, err := creator.Create()

						Expect(err).To(Equal(expectedErr))
					})
				})
			})

			Context("when the Override parameter is turned off", func() {
				BeforeEach(func() {
					params.Override = false
				})

				It("returns an error", func() {
					_, err := creator.Create()
					Expect(err).To(MatchError(fmt.Errorf("Release '%s' with version '%s' already exists.", productSlug, releaseVersion)))
				})
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

		Context("When copying metadata", func() {
			BeforeEach(func() {
				copyMetadata = true
			})

			It("constructs the release", func() {
				r, err := creator.Create()
				Expect(err).NotTo(HaveOccurred())

				Expect(r).To(Equal(pivnet.Release{ID: 1337}))

				Expect(pivnetClient.EULAsCallCount()).To(Equal(1))

				Expect(pivnetClient.ReleasesForProductSlugArgsForCall(0)).To(Equal(productSlug))

				Expect(pivnetClient.CreateReleaseArgsForCall(0)).To(Equal(pivnet.CreateReleaseConfig{
					ProductSlug:     productSlug,
					ReleaseType:     string(releaseType),
					EULASlug:        eulaSlug,
					Version:         releaseVersion,
					Description:     "wow, a description",
					ReleaseNotesURL: "some-url",
					ReleaseDate:     "1/17/2016",
					Controlled:      true,
					CopyMetadata:    copyMetadata,
				}))
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
