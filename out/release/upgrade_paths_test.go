package release_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-resource/v2/metadata"
	"github.com/pivotal-cf/pivnet-resource/v2/out/release"
	"github.com/pivotal-cf/pivnet-resource/v2/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseUpgradePathsAdder", func() {
	Describe("AddReleaseUpgradePaths", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseUpgradePathsAdderClient
			fakeFilter   *releasefakes.FakeFilter

			existingReleases []pivnet.Release
			filteredReleases []pivnet.Release

			existingReleasesErr error
			filterErr           error

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseUpgradePathsAdder release.ReleaseUpgradePathsAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseUpgradePathsAdderClient{}
			fakeFilter = &releasefakes.FakeFilter{}

			existingReleases = []pivnet.Release{
				{
					ID:      1234,
					Version: "1.2.3",
				},
				{
					ID:      1235,
					Version: "1.2.4",
				},
				{
					ID:      2345,
					Version: "2.3.4",
				},
				{
					ID:      3456,
					Version: "3.4.5",
				},
			}

			filteredReleases = []pivnet.Release{
				existingReleases[0],
				existingReleases[1],
			}

			productSlug = "some-product-slug"

			pivnetRelease = existingReleases[2]

			mdata = metadata.Metadata{
				Release: &metadata.Release{
					Version: "some-version",
				},
				ProductFiles: []metadata.ProductFile{},
				UpgradePaths: []metadata.UpgradePath{
					{},
				},
			}

			existingReleasesErr = nil
			filterErr = nil
		})

		JustBeforeEach(func() {
			releaseUpgradePathsAdder = release.NewReleaseUpgradePathsAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
				fakeFilter,
			)

			pivnetClient.ReleasesForProductSlugReturns(existingReleases, existingReleasesErr)
			fakeFilter.ReleasesByVersionReturns(filteredReleases, filterErr)
		})

		Describe("upgrade path via ID", func() {
			BeforeEach(func() {
				mdata.UpgradePaths[0].ID = existingReleases[0].ID
				mdata.UpgradePaths[0].Version = ""
			})

			It("adds the upgrade paths", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(1))
				invokedProductSlug, invokedReleaseID, invokedPreviousReleaseID :=
					pivnetClient.AddReleaseUpgradePathArgsForCall(0)
				Expect(invokedProductSlug).To(Equal(productSlug))
				Expect(invokedReleaseID).To(Equal(pivnetRelease.ID))
				Expect(invokedPreviousReleaseID).To(Equal(existingReleases[0].ID))
			})

			Context("when provided ID does not match any existing release", func() {
				BeforeEach(func() {
					mdata.UpgradePaths[0].ID = 19283
				})

				It("returns an error", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(MatchRegexp("No releases found for id: '%d'", mdata.UpgradePaths[0].ID))
				})
			})

			Context("when release matches upgrade path", func() {
				BeforeEach(func() {
					mdata.UpgradePaths[0].ID = pivnetRelease.ID
				})

				It("does not attempt to add itself as an upgrade path", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(0))
				})
			})
		})

		Describe("upgrade path via version", func() {
			BeforeEach(func() {
				mdata.UpgradePaths[0].ID = 0
				mdata.UpgradePaths[0].Version = "1.2.*"
			})

			It("filters all the releases by the provided version", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(2))
				Expect(fakeFilter.ReleasesByVersionCallCount()).To(Equal(1))
			})

			Context("when filtering returns the same upgrade path for different filters", func() {
				BeforeEach(func() {
					mdata.UpgradePaths = []metadata.UpgradePath{
						{
							Version: "1.2.*",
						},
						{
							Version: "this will also match the same releases",
						},
					}
				})

				It("only attempts to add each unique upgrade path once", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(2))
					Expect(fakeFilter.ReleasesByVersionCallCount()).To(Equal(2))
				})
			})

			Context("when filtering releases returns an error", func() {
				BeforeEach(func() {
					filterErr = errors.New("filter err")
				})

				It("returns an error", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(filterErr))
				})
			})

			Context("when filtered releases are empty", func() {
				BeforeEach(func() {
					filteredReleases = []pivnet.Release{}
				})

				It("returns an error", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(MatchRegexp("No releases found for version: '%s'", mdata.UpgradePaths[0].Version))
				})
			})
		})

		Context("when previous release version is empty and ID is 0", func() {
			BeforeEach(func() {
				mdata.UpgradePaths[0].Version = ""
				mdata.UpgradePaths[0].ID = 0
			})

			It("returns an error", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("upgrade_paths[0]"))
			})
		})

		Context("when getting all releases returns an error ", func() {
			BeforeEach(func() {
				existingReleasesErr = errors.New("existing releases err")
			})

			It("returns an error", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(existingReleasesErr))
			})
		})

		Context("when adding upgrade path returns an error ", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				mdata.UpgradePaths[0].ID = existingReleases[0].ID

				expectedErr = fmt.Errorf("boom")
				pivnetClient.AddReleaseUpgradePathReturns(expectedErr)
			})

			It("returns an error", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
