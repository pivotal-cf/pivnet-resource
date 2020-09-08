package release_test

import (
	"log"
	"time"

	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseArtifactReferencesAdder", func() {
	Describe("AddReleaseArtifactReferences", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseArtifactReferencesAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseArtifactReferencesAdder release.ReleaseArtifactReferencesAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseArtifactReferencesAdderClient{}

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

			pivnetClient.AddArtifactReferenceReturns(nil)
		})

		JustBeforeEach(func() {
			releaseArtifactReferencesAdder = release.NewReleaseArtifactReferencesAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
				time.Millisecond,
				time.Second,
			)
		})

		Context("when release ArtifactReferences are provided", func() {
			var (
				ref1 pivnet.ArtifactReference
				ref2 pivnet.ArtifactReference
			)
			BeforeEach(func() {
				mdata.ArtifactReferences = []metadata.ArtifactReference{
					{
						ID:   9876,
						Name: "my-difficult-artifact",
					},
					{
						ID:           1234,
						Name:         "new-artifact-reference",
						ArtifactPath: "my/path:123",
						Digest:       "sha256:mydigest",
					},
				}

				ref1 = pivnet.ArtifactReference{
					ID:                9876,
					ReplicationStatus: pivnet.Complete,
					Name:              "my-difficult-artifact",
				}
				ref2 = pivnet.ArtifactReference{
					ID:                1234,
					ReplicationStatus: pivnet.Complete,
					Name:              "new-artifact-reference",
				}
				pivnetClient.GetArtifactReferenceStub = func(slug string, id int) (pivnet.ArtifactReference, error) {
					if id == 9876 {
						return ref1, nil
					} else if id == 1234 {
						return ref2, nil
					} else {
						return pivnet.ArtifactReference{}, fmt.Errorf("missing stub for artifact %d", id)
					}
				}
			})

			It("adds the ArtifactReferences", func() {
				err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddArtifactReferenceCallCount()).To(Equal(2))

				Expect(pivnetClient.GetArtifactReferenceCallCount()).To(Equal(2))
				_, artifactReferenceID := pivnetClient.GetArtifactReferenceArgsForCall(0)
				Expect(artifactReferenceID).To(Equal(9876))
				_, artifactReferenceID = pivnetClient.GetArtifactReferenceArgsForCall(1)
				Expect(artifactReferenceID).To(Equal(1234))
			})

			Context("when the artifact reference ID is set to 0", func() {
				BeforeEach(func() {
					pivnetClient.CreateArtifactReferenceReturns(ref2, nil)
					mdata.ArtifactReferences[1].ID = 0
				})

				It("creates a new artifact reference", func() {
					err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddArtifactReferenceCallCount()).To(Equal(2))
					Expect(pivnetClient.CreateArtifactReferenceCallCount()).To(Equal(1))
				})

				Context("when name, artifactPath, and digest are the same as an existing artifact reference", func() {
					BeforeEach(func() {
						pivnetClient.ArtifactReferencesReturns(
							[]pivnet.ArtifactReference{
								{
									ID:           1234,
									Name:         "new-artifact-reference",
									ArtifactPath: "my/path:123",
									Digest:       "sha256:mydigest",
								},
							}, nil)
					})

					It("does uses the existing artifact reference", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())
						Expect(pivnetClient.CreateArtifactReferenceCallCount()).To(Equal(0))
						Expect(pivnetClient.AddArtifactReferenceCallCount()).To(Equal(2))
						_, _, artifactReferenceID := pivnetClient.AddArtifactReferenceArgsForCall(0)
						Expect(artifactReferenceID).To(Equal(9876))
						_, _, artifactReferenceID = pivnetClient.AddArtifactReferenceArgsForCall(1)
						Expect(artifactReferenceID).To(Equal(1234))
					})
				})
				Context("when creating the artifact reference returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some artifact reference error")
						pivnetClient.CreateArtifactReferenceReturns(pivnet.ArtifactReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when artifact replication is in progress", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetArtifactReferenceReturnsOnCall(0, ref1, nil)
						ref1.ReplicationStatus = pivnet.Complete
						pivnetClient.GetArtifactReferenceReturnsOnCall(1, ref1, nil)

						ref2.ReplicationStatus = pivnet.InProgress
						pivnetClient.GetArtifactReferenceReturnsOnCall(2, ref2, nil)
						ref2.ReplicationStatus = pivnet.Complete
						pivnetClient.GetArtifactReferenceReturnsOnCall(3, ref2, nil)
					})

					It("waits for replication to complete", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).NotTo(HaveOccurred())

						Expect(pivnetClient.GetArtifactReferenceCallCount()).To(Equal(4))
						_, artifactReferenceID := pivnetClient.GetArtifactReferenceArgsForCall(0)
						Expect(artifactReferenceID).To(Equal(9876))
						_, artifactReferenceID = pivnetClient.GetArtifactReferenceArgsForCall(1)
						Expect(artifactReferenceID).To(Equal(9876))
						_, artifactReferenceID = pivnetClient.GetArtifactReferenceArgsForCall(2)
						Expect(artifactReferenceID).To(Equal(1234))
						_, artifactReferenceID = pivnetClient.GetArtifactReferenceArgsForCall(3)
						Expect(artifactReferenceID).To(Equal(1234))
					})
				})

				Context("when artifact replication fails", func() {
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.FailedToReplicate
					})

					It("returns an error", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						expectedErr := fmt.Errorf("artifact reference with name my-difficult-artifact failed to replicate")
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetArtifactReferenceCallCount()).To(Equal(1))
						_, artifactReferenceID := pivnetClient.GetArtifactReferenceArgsForCall(0)
						Expect(artifactReferenceID).To(Equal(9876))
					})
				})

				Context("when checking artifact replication fails", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						expectedErr = fmt.Errorf("some network flake")
						pivnetClient.GetArtifactReferenceReturnsOnCall(0, pivnet.ArtifactReference{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(expectedErr))

						Expect(pivnetClient.GetArtifactReferenceCallCount()).To(Equal(1))
						_, artifactReferenceID := pivnetClient.GetArtifactReferenceArgsForCall(0)
						Expect(artifactReferenceID).To(Equal(9876))
					})
				})

				Context("when checking artifact replication times out", func() {
					var (
						expectedErr error
					)
					BeforeEach(func() {
						ref1.ReplicationStatus = pivnet.InProgress
						expectedErr = fmt.Errorf("timed out replicating artifact reference with name: my-difficult-artifact")
					})

					It("returns an error", func() {
						err := releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})
			})
		})
	})
})
