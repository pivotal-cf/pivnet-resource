package release_test

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFinalizer", func() {
	Describe("Finalize", func() {
		var (
			fakePivnet *releasefakes.FinalizerClient
			l          *log.Logger
			params     concourse.OutParams

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseErr error

			finalizer release.ReleaseFinalizer
		)

		BeforeEach(func() {
			fakePivnet = &releasefakes.FinalizerClient{}
			l = log.New(ioutil.Discard, "it doesn't matter", 0)

			params = concourse.OutParams{}

			productSlug = "some-product-slug"

			pivnetRelease = pivnet.Release{
				Availability: "Admins Only",
				ID:           1337,
				Version:      "some-version",
				EULA: &pivnet.EULA{
					Slug: "a_eula_slug",
				},
				UpdatedAt: "some-new-time",
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
				l,
				params,
				mdata,
				"/some/sources/dir",
				productSlug,
			)

			fakePivnet.GetReleaseReturns(pivnetRelease, releaseErr)
		})

		It("returns a final concourse out response", func() {
			response, err := finalizer.Finalize(productSlug, pivnetRelease.Version)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Version).To(Equal(concourse.Version{
				ProductVersion: "some-version#some-new-time",
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
