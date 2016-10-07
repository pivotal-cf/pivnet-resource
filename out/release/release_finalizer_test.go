package release_test

import (
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
			pivnetClient *releasefakes.UpdateClient
			params       concourse.OutParams

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			finalizer release.ReleaseFinalizer
		)

		BeforeEach(func() {
			pivnetClient = &releasefakes.UpdateClient{}

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

			pivnetClient.ReleaseETagReturns("a-diff-etag", nil)
		})

		JustBeforeEach(func() {
			finalizer = release.NewFinalizer(
				pivnetClient,
				params,
				mdata,
				"/some/sources/dir",
				productSlug,
			)
		})

		It("returns a final concourse out response", func() {
			response, err := finalizer.Finalize(pivnetRelease)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Version).To(Equal(concourse.Version{
				ProductVersion: "some-version#a-diff-etag",
			}))

			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "version", Value: "some-version"}))
			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "controlled", Value: "false"}))
			Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "eula_slug", Value: "a_eula_slug"}))
		})
	})
})
