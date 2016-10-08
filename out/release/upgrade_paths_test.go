package release_test

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseUpgradePathsAdder", func() {
	Describe("AddReleaseUpgradePaths", func() {
		var (
			l *log.Logger

			pivnetClient *releasefakes.ReleaseUpgradePathsAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseUpgradePathsAdder release.ReleaseUpgradePathsAdder
		)

		BeforeEach(func() {
			pivnetClient = &releasefakes.ReleaseUpgradePathsAdderClient{}
			l = log.New(ioutil.Discard, "it doesn't matter", 0)

			productSlug = "some-product-slug"

			pivnetRelease = pivnet.Release{
				Availability: "some-value",
				ID:           1337,
				Version:      "some-version",
				EULA: &pivnet.EULA{
					Slug: "a_eula_slug",
				},
			}

			mdata = metadata.Metadata{
				Release: &metadata.Release{
					Availability: "some-value",
					Version:      "some-version",
					EULASlug:     "a_eula_slug",
				},
				ProductFiles: []metadata.ProductFile{},
			}

			pivnetClient.AddReleaseUpgradePathReturns(nil)
		})

		JustBeforeEach(func() {
			releaseUpgradePathsAdder = release.NewReleaseUpgradePathsAdder(
				l,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when release upgrade paths are provided", func() {
			BeforeEach(func() {
				mdata.UpgradePaths = []metadata.UpgradePath{
					{
						ID:      9876,
						Version: "some-previous-release-version",
					},
					{
						ID:      8765,
						Version: "some-other-previous-release-version",
					},
				}
			})

			It("adds the upgrade paths", func() {
				err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(2))
			})

			Context("when the previous release ID is zero", func() {
				BeforeEach(func() {
					mdata.UpgradePaths[1].ID = 0

					previousRelease := pivnet.Release{
						ID: 9876,
					}

					pivnetClient.GetReleaseReturns(previousRelease, nil)
				})

				It("obtains the previous release from the version and product slug", func() {
					err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddReleaseUpgradePathCallCount()).To(Equal(2))
					Expect(pivnetClient.GetReleaseCallCount()).To(Equal(1))
				})

				Context("when obtaining the previous release returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some release error")
						pivnetClient.GetReleaseReturns(pivnet.Release{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when previous release version is empty", func() {
					BeforeEach(func() {
						mdata.UpgradePaths[1].Version = ""
					})

					It("returns an error", func() {
						err := releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err.Error()).To(ContainSubstring("upgrade_paths[1]"))
					})
				})
			})

			Context("when adding dependency returns an error ", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
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
})
