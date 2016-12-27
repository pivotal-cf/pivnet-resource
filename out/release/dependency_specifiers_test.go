package release_test

import (
	"fmt"
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DependencySpecifiersCreator", func() {
	Describe("CreateDependencySpecifiers", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.DependencySpecifiersCreatorClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			dependencySpecifiersCreator release.DependencySpecifiersCreator
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.DependencySpecifiersCreatorClient{}

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

			pivnetClient.CreateDependencySpecifierReturns(pivnet.DependencySpecifier{}, nil)
		})

		JustBeforeEach(func() {
			dependencySpecifiersCreator = release.NewDependencySpecifiersCreator(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when dependency specifiers are provided", func() {
			BeforeEach(func() {
				mdata.DependencySpecifiers = []metadata.DependencySpecifier{
					{
						ID:          9876,
						Specifier:   "1.2.*",
						ProductSlug: "some-dependent-product",
					},
					{
						ID:          8765,
						Specifier:   "2.3.*",
						ProductSlug: "some-other-dependent-product",
					},
				}
			})

			It("creates the dependencies", func() {
				err := dependencySpecifiersCreator.CreateDependencySpecifiers(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.CreateDependencySpecifierCallCount()).To(Equal(2))
			})

			Context("when creating dependency specifier returns an error ", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = fmt.Errorf("boom")
					pivnetClient.CreateDependencySpecifierReturns(pivnet.DependencySpecifier{}, expectedErr)
				})

				It("returns an error", func() {
					err := dependencySpecifiersCreator.CreateDependencySpecifiers(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
