package release_test

import (
	"errors"
	"log"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
	"github.com/pivotal-cf/pivnet-resource/v3/out/release"
	"github.com/pivotal-cf/pivnet-resource/v3/out/release/releasefakes"
	"github.com/pivotal-cf/pivnet-resource/v3/versions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFinalizer", func() {
	Describe("Finalize", func() {
		var (
			fakeLogger logger.Logger

			fakePivnet *releasefakes.FinalizerClient
			params     concourse.OutParams

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseErr error

			finalizer release.ReleaseFinalizer
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			fakePivnet = &releasefakes.FinalizerClient{}

			params = concourse.OutParams{}

			productSlug = "some-product-slug"

			pivnetRelease = pivnet.Release{
				Availability: "Admins Only",
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

			releaseErr = nil
		})

		JustBeforeEach(func() {
			finalizer = release.NewFinalizer(
				fakePivnet,
				fakeLogger,
				params,
				mdata,
				"/some/sources/dir",
				productSlug,
			)

			fakePivnet.GetReleaseReturns(pivnetRelease, releaseErr)
			// Stub product files for file-based fingerprint (TNZ-22056)
			fakePivnet.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				{ID: 1, ReleasedAt: "some-new-time"},
			}, nil)
			fakePivnet.FileGroupsForReleaseReturns([]pivnet.FileGroup{}, nil)
		})

		It("returns a final concourse out response", func() {
			response, err := finalizer.Finalize(productSlug, pivnetRelease.Version)
			Expect(err).NotTo(HaveOccurred())

			expectedFingerprint := versions.FingerprintFromFileMetadata([]versions.FileMetadata{
				{ID: 1, ReleasedAt: "some-new-time"},
			})
			expectedVersion, _ := versions.CombineVersionAndFingerprint("some-version", expectedFingerprint)
			Expect(response.Version).To(Equal(concourse.Version{
				ProductVersion: expectedVersion,
			}))

			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "version", Value: "some-version"}))
			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "controlled", Value: "false"}))
			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "eula_slug", Value: "a_eula_slug"}))
		})

		Context("when getting the release returns an error", func() {
			BeforeEach(func() {
				releaseErr = errors.New("release error")
			})

			It("forwards the error", func() {
				_, err := finalizer.Finalize(productSlug, pivnetRelease.Version)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(releaseErr))
			})
		})
	})
})
