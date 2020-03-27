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

var _ = Describe("ReleaseImageReferencesAdder", func() {
	Describe("AddReleaseImageReferences", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseImageReferencesAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseImageReferencesAdder release.ReleaseImageReferencesAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseImageReferencesAdderClient{}

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

			pivnetClient.AddImageReferenceReturns(nil)
		})

		JustBeforeEach(func() {
			releaseImageReferencesAdder = release.NewReleaseImageReferencesAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
				time.Millisecond,
				time.Second,
			)
		})

		Context("when release ImageReferences are provided", func() {
			var (
				ref1 pivnet.ImageReference
				ref2 pivnet.ImageReference
			)
			BeforeEach(func() {
				mdata.ImageReferences = []metadata.ImageReference{
					{
						ID: 9876,
						Name: "my-difficult-image",
					},
					{
						ID:        1234,
						Name:      "new-image-reference",
						ImagePath: "my/path:123",
						Digest:    "sha256:mydigest",
					},
				}

				ref1 = pivnet.ImageReference{
					ID:                 9876,
					ReplicationStatus:  pivnet.Complete,
					Name: "my-difficult-image",
				}
				ref2 = pivnet.ImageReference{
					ID:        1234,
					ReplicationStatus: pivnet.Complete,
					Name:      "new-image-reference",
				}
				pivnetClient.GetImageReferenceStub = func(slug string, id int) (pivnet.ImageReference, error) {
					if id == 9876 {
						return ref1, nil
					} else if id == 1234 {
						return ref2, nil
					} else {
						return pivnet.ImageReference{}, fmt.Errorf("missing stub for image %d", id)
					}
				}
			})

			It("adds the ImageReferences", func() {
				err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddImageReferenceCallCount()).To(Equal(2))

				Expect(pivnetClient.GetImageReferenceCallCount()).To(Equal(2))
				_, imageReferenceID := pivnetClient.GetImageReferenceArgsForCall(0)
				Expect(imageReferenceID).To(Equal(9876))
				_, imageReferenceID = pivnetClient.GetImageReferenceArgsForCall(1)
				Expect(imageReferenceID).To(Equal(1234))
			})

			Context("when the image reference ID is set to 0", func() {
				BeforeEach(func() {
					pivnetClient.CreateImageReferenceReturns(ref2, nil)
					mdata.ImageReferences[1].ID = 0
				})

				It("creates a new image reference", func() {
					err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddImageReferenceCallCount()).To(Equal(2))
					Expect(pivnetClient.CreateImageReferenceCallCount()).To(Equal(1))
				})

				Context("when name, imagePath, and digest are the same as an existing image reference", func() {
					BeforeEach(func() {
						pivnetClient.ImageReferencesReturns(
							[]pivnet.ImageReference{
								{
									ID:        1234,
									Name:      "new-image-reference",
									ImagePath: "my/path:123",
									Digest:    "sha256:mydigest",
								},
							}, nil)
					})

					It("does uses the existing image reference", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())
						Expect(pivnetClient.CreateImageReferenceCallCount()).To(Equal(0))
						Expect(pivnetClient.AddImageReferenceCallCount()).To(Equal(2))
						_, _, imageReferenceID := pivnetClient.AddImageReferenceArgsForCall(0)
						Expect(imageReferenceID).To(Equal(9876))
						_, _, imageReferenceID = pivnetClient.AddImageReferenceArgsForCall(1)
						Expect(imageReferenceID).To(Equal(1234))
					})
				})
				Context("when creating the image reference returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some image reference error")
						pivnetClient.CreateImageReferenceReturns(pivnet.ImageReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when image replication is in progress", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetImageReferenceReturnsOnCall(0, ref1, nil)
						ref1.ReplicationStatus = pivnet.Complete
						pivnetClient.GetImageReferenceReturnsOnCall(1, ref1, nil)

						ref2.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetImageReferenceReturnsOnCall(2, ref2, nil)
						ref2.ReplicationStatus = pivnet.Complete
						pivnetClient.GetImageReferenceReturnsOnCall(3, ref2, nil)
					})

					It("waits for replication to complete", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())

						Expect(pivnetClient.GetImageReferenceCallCount()).To(Equal(4))
						_, imageReferenceID := pivnetClient.GetImageReferenceArgsForCall(0)
						Expect(imageReferenceID).To(Equal(9876))
						_, imageReferenceID = pivnetClient.GetImageReferenceArgsForCall(1)
						Expect(imageReferenceID).To(Equal(9876))
						_, imageReferenceID = pivnetClient.GetImageReferenceArgsForCall(2)
						Expect(imageReferenceID).To(Equal(1234))
						_, imageReferenceID = pivnetClient.GetImageReferenceArgsForCall(3)
						Expect(imageReferenceID).To(Equal(1234))
					})
				})

				Context("when image replication fails", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.FailedToReplicate
					})

					It("returns an error", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						expectedErr := fmt.Errorf("image reference with name my-difficult-image failed to replicate")
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetImageReferenceCallCount()).To(Equal(1))
						_, imageReferenceID := pivnetClient.GetImageReferenceArgsForCall(0)
						Expect(imageReferenceID).To(Equal(9876))
					})
				})

				Context("when checking image replication fails", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						expectedErr = fmt.Errorf("some network flake")
						pivnetClient.GetImageReferenceReturnsOnCall(0, pivnet.ImageReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetImageReferenceCallCount()).To(Equal(1))
						_, imageReferenceID := pivnetClient.GetImageReferenceArgsForCall(0)
						Expect(imageReferenceID).To(Equal(9876))
					})
				})

				Context("when checking image replication times out", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						expectedErr = fmt.Errorf("timed out replicating image reference with name: my-difficult-image")
					})

					It("returns an error", func() {
						err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})
			})
		})
	})
})
