package release_test

import (
	"log"

	"github.com/pivotal-cf/go-pivnet/v3"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
	"github.com/pivotal-cf/go-pivnet/v3/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseHelmChartReferencesAdder", func() {
	Describe("AddReleaseHelmChartReferences", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseHelmChartReferencesAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseHelmChartReferencesAdder release.ReleaseHelmChartReferencesAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseHelmChartReferencesAdderClient{}

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

			pivnetClient.AddHelmChartReferenceReturns(nil)
		})

		JustBeforeEach(func() {
			releaseHelmChartReferencesAdder = release.NewReleaseHelmChartReferencesAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when release HelmChartReferences are provided", func() {
			BeforeEach(func() {
				mdata.HelmChartReferences = []metadata.HelmChartReference{
					{
						ID: 9876,
					},
					{
						ID:      1234,
						Name:    "new-helm-chart-reference",
						Version: "1.2.3",
					},
				}
			})

			It("adds the HelmChartReferences", func() {
				err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddHelmChartReferenceCallCount()).To(Equal(2))
			})

			Context("when the helm chart reference ID is set to 0", func() {
				BeforeEach(func() {
					mdata.HelmChartReferences[1].ID = 0
				})

				It("creates a new helm chart reference", func() {
					err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddHelmChartReferenceCallCount()).To(Equal(2))
					Expect(pivnetClient.CreateHelmChartReferenceCallCount()).To(Equal(1))
				})

				Context("when name and version are the same as an existing helm chart reference", func() {
					BeforeEach(func() {
						pivnetClient.HelmChartReferencesReturns(
							[]pivnet.HelmChartReference{
								{
									ID:      1234,
									Name:    "new-helm-chart-reference",
									Version: "1.2.3",
								},
							}, nil)
					})

					It("does uses the existing helm chart reference", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())
						Expect(pivnetClient.CreateHelmChartReferenceCallCount()).To(Equal(0))
						Expect(pivnetClient.AddHelmChartReferenceCallCount()).To(Equal(2))
						_, _, helmChartReferenceID := pivnetClient.AddHelmChartReferenceArgsForCall(0)
						Expect(helmChartReferenceID).To(Equal(9876))
						_, _, helmChartReferenceID = pivnetClient.AddHelmChartReferenceArgsForCall(1)
						Expect(helmChartReferenceID).To(Equal(1234))
					})
				})
				Context("when creating the helm chart reference returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some helm chart reference error")
						pivnetClient.CreateHelmChartReferenceReturns(pivnet.HelmChartReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})
			})
		})
	})
})
