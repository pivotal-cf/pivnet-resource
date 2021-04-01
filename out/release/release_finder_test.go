package release_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/pivnet-resource/v3/out/release"
	"github.com/pivotal-cf/pivnet-resource/v3/out/release/releasefakes"

	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFinder", func() {
	var (
		pivnetClient *releasefakes.FinderClient
		finder       release.ReleaseFinder
		releaseID    int
		pivnetRelease pivnet.Release
		finderErr error
		productSlug string
	)
	BeforeEach(func() {
		pivnetClient = &releasefakes.FinderClient{}
		finderErr = nil
		productSlug = "special-product"
	})

	Describe("Find", func() {

		Context("", func() {

		})
		BeforeEach(func() {
			finder = release.NewReleaseFinder(
				pivnetClient,
				productSlug,
			)
			releaseID = 123
			pivnetRelease = pivnet.Release{
				ID: releaseID,
			}

			pivnetClient.FindReleaseReturns(pivnetRelease, finderErr)
		})

		It("finds a release", func() {
			r, err := finder.Find(releaseID)
			Expect(err).NotTo(HaveOccurred())

			slug, id := pivnetClient.FindReleaseArgsForCall(0)
			Expect(slug).To(Equal(productSlug))
			Expect(id).To(Equal(releaseID))

			Expect(r).To(Equal(pivnet.Release{ID: releaseID}))
		})
	})
})
