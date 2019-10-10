package release_test

import (
	"log"

	"github.com/pivotal-cf/go-pivnet/v2"
	"github.com/pivotal-cf/go-pivnet/v2/logger"
	"github.com/pivotal-cf/go-pivnet/v2/logshim"
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
			)
		})

		Context("when release ImageReferences are provided", func() {
			BeforeEach(func() {
				mdata.ImageReferences = []metadata.ImageReference{
					{
						ID: 9876,
					},
					{
						ID:        1234,
						Name:      "new-image-reference",
						ImagePath: "my/path:123",
						Digest:    "sha256:mydigest",
					},
				}
			})

			It("adds the ImageReferences", func() {
				err := releaseImageReferencesAdder.AddReleaseImageReferences(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddImageReferenceCallCount()).To(Equal(2))
			})

			Context("when the image reference ID is set to 0", func() {
				BeforeEach(func() {
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
			})
		})
	})
})
