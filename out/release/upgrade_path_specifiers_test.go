package release_test

import (
	"fmt"
	"log"

	"github.com/pivotal-cf/go-pivnet/v3"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
	"github.com/pivotal-cf/go-pivnet/v3/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradePathSpecifiersCreator", func() {
	Describe("CreateUpgradePathSpecifiers", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.UpgradePathSpecifiersCreatorClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			upgradePathSpecifiersCreator release.UpgradePathSpecifiersCreator
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.UpgradePathSpecifiersCreatorClient{}

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

			pivnetClient.CreateUpgradePathSpecifierReturns(pivnet.UpgradePathSpecifier{}, nil)
		})

		JustBeforeEach(func() {
			upgradePathSpecifiersCreator = release.NewUpgradePathSpecifiersCreator(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when upgrade path specifiers are provided", func() {
			BeforeEach(func() {
				mdata.UpgradePathSpecifiers = []metadata.UpgradePathSpecifier{
					{
						ID:        9876,
						Specifier: "1.2.*",
					},
					{
						ID:        8765,
						Specifier: "2.3.*",
					},
				}
			})

			It("creates the upgrade paths", func() {
				err := upgradePathSpecifiersCreator.CreateUpgradePathSpecifiers(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.CreateUpgradePathSpecifierCallCount()).To(Equal(2))
			})

			Context("when creating upgrade path specifier returns an error ", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = fmt.Errorf("boom")
					pivnetClient.CreateUpgradePathSpecifierReturns(pivnet.UpgradePathSpecifier{}, expectedErr)
				})

				It("returns an error", func() {
					err := upgradePathSpecifiersCreator.CreateUpgradePathSpecifiers(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
