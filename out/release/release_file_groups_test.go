package release_test

import (
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFileGroupsAdder", func() {
	Describe("AddReleaseFileGroups", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseFileGroupsAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseFileGroupsAdder release.ReleaseFileGroupsAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseFileGroupsAdderClient{}

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

			pivnetClient.AddFileGroupReturns(nil)
		})

		JustBeforeEach(func() {
			releaseFileGroupsAdder = release.NewReleaseFileGroupsAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when release FileGroups are provided", func() {
			BeforeEach(func() {
				mdata.FileGroups = []metadata.FileGroup{
					{
						ID: 9876,
					},
					{
						ID:   1234,
						Name: "new-file-group",
					},
				}
			})

			It("adds the FileGroups", func() {
				err := releaseFileGroupsAdder.AddReleaseFileGroups(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddFileGroupCallCount()).To(Equal(2))
			})

			Context("when the file group ID is set to 0", func() {
				BeforeEach(func() {
					mdata.FileGroups[1].ID = 0
				})

				It("creates a new file group", func() {
					err := releaseFileGroupsAdder.AddReleaseFileGroups(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddFileGroupCallCount()).To(Equal(2))
					Expect(pivnetClient.CreateFileGroupCallCount()).To(Equal(1))
				})

				Context("when creating the file group returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some file group error")
						pivnetClient.CreateFileGroupReturns(pivnet.FileGroup{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseFileGroupsAdder.AddReleaseFileGroups(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})
			})
		})
	})
})
