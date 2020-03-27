package release_test

import (
	"log"
	"time"

	"github.com/pivotal-cf/go-pivnet/v4"
	"github.com/pivotal-cf/go-pivnet/v4/logger"
	"github.com/pivotal-cf/go-pivnet/v4/logshim"
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
				time.Millisecond,
				time.Second,
			)
		})

		Context("when release HelmChartReferences are provided", func() {
			var (
				ref1 pivnet.HelmChartReference
				ref2 pivnet.HelmChartReference
			)
			BeforeEach(func() {
				mdata.HelmChartReferences = []metadata.HelmChartReference{
					{
						ID: 9876,
						Name: "my-difficult-helm-chart",
					},
					{
						ID:      1234,
						Name:    "new-helm-chart-reference",
						Version: "1.2.3",
					},
				}

				ref1 = pivnet.HelmChartReference{
					ID:                 9876,
					ReplicationStatus:  pivnet.Complete,
					Name: "my-difficult-helm-chart",
				}
				ref2 = pivnet.HelmChartReference{
					ID:        1234,
					ReplicationStatus: pivnet.Complete,
					Name:    "new-helm-chart-reference",
				}
				pivnetClient.GetHelmChartReferenceStub = func(slug string, id int) (pivnet.HelmChartReference, error) {
					if id == 9876 {
						return ref1, nil
					} else if id == 1234 {
						return ref2, nil
					} else {
						return pivnet.HelmChartReference{}, fmt.Errorf("missing stub for chart %d", id)
					}
				}
			})

			It("adds the HelmChartReferences", func() {
				err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddHelmChartReferenceCallCount()).To(Equal(2))

				Expect(pivnetClient.GetHelmChartReferenceCallCount()).To(Equal(2))
				_, helmChartReferenceID := pivnetClient.GetHelmChartReferenceArgsForCall(0)
				Expect(helmChartReferenceID).To(Equal(9876))
				_, helmChartReferenceID = pivnetClient.GetHelmChartReferenceArgsForCall(1)
				Expect(helmChartReferenceID).To(Equal(1234))
			})

			Context("when the helm chart reference ID is set to 0", func() {
				BeforeEach(func() {
					pivnetClient.CreateHelmChartReferenceReturns(ref2, nil)
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

				Context("when helmChart replication is in progress", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetHelmChartReferenceReturnsOnCall(0, ref1, nil)
						ref1.ReplicationStatus = pivnet.Complete
						pivnetClient.GetHelmChartReferenceReturnsOnCall(1, ref1, nil)


						ref2.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetHelmChartReferenceReturnsOnCall(2, ref2, nil)
						ref2.ReplicationStatus = pivnet.Complete
						pivnetClient.GetHelmChartReferenceReturnsOnCall(3, ref2, nil)
					})

					It("waits for replication to complete", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())

						Expect(pivnetClient.GetHelmChartReferenceCallCount()).To(Equal(4))
						_, helmChartReferenceID := pivnetClient.GetHelmChartReferenceArgsForCall(0)
						Expect(helmChartReferenceID).To(Equal(9876))
						_, helmChartReferenceID = pivnetClient.GetHelmChartReferenceArgsForCall(1)
						Expect(helmChartReferenceID).To(Equal(9876))
						_, helmChartReferenceID = pivnetClient.GetHelmChartReferenceArgsForCall(2)
						Expect(helmChartReferenceID).To(Equal(1234))
						_, helmChartReferenceID = pivnetClient.GetHelmChartReferenceArgsForCall(3)
						Expect(helmChartReferenceID).To(Equal(1234))
					})
				})

				Context("when helm chart replication fails", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.FailedToReplicate
					})

					It("returns an error", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						expectedErr := fmt.Errorf("helm chart reference with name my-difficult-helm-chart failed to replicate")
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetHelmChartReferenceCallCount()).To(Equal(1))
						_, helmChartReferenceID := pivnetClient.GetHelmChartReferenceArgsForCall(0)
						Expect(helmChartReferenceID).To(Equal(9876))
					})
				})

				Context("when checking helm chart replication fails", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						expectedErr = fmt.Errorf("some network flake")
						pivnetClient.GetHelmChartReferenceReturnsOnCall(0, pivnet.HelmChartReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetHelmChartReferenceCallCount()).To(Equal(1))
						_, helmChartReferenceID := pivnetClient.GetHelmChartReferenceArgsForCall(0)
						Expect(helmChartReferenceID).To(Equal(9876))
					})
				})

				Context("when checking image replication times out", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						expectedErr = fmt.Errorf("timed out replicating helm chart reference with name: my-difficult-helm-chart")
					})

					It("returns an error", func() {
						err := releaseHelmChartReferencesAdder.AddReleaseHelmChartReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})
			})
		})
	})
})
