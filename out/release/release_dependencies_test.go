package release_test

import (
	"fmt"
	"log"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-resource/v2/metadata"
	"github.com/pivotal-cf/pivnet-resource/v2/out/release"
	"github.com/pivotal-cf/pivnet-resource/v2/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseDependenciesAdder", func() {
	Describe("AddReleaseDependencies", func() {
		var (
			fakeLogger logger.Logger

			pivnetClient *releasefakes.ReleaseDependenciesAdderClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			releaseDependenciesAdder release.ReleaseDependenciesAdder
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			pivnetClient = &releasefakes.ReleaseDependenciesAdderClient{}

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

			pivnetClient.AddReleaseDependencyReturns(nil)
		})

		JustBeforeEach(func() {
			releaseDependenciesAdder = release.NewReleaseDependenciesAdder(
				fakeLogger,
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		Context("when release dependencies are provided", func() {
			BeforeEach(func() {
				mdata.Dependencies = []metadata.Dependency{
					{
						Release: metadata.DependentRelease{
							ID:      9876,
							Version: "some-dependent-release-version",
							Product: metadata.Product{
								Slug: "some-dependent-product",
							},
						},
					},
					{
						Release: metadata.DependentRelease{
							ID:      8765,
							Version: "some-other-dependent-release-version",
							Product: metadata.Product{
								Slug: "some-other-dependent-product",
							},
						},
					},
				}
			})

			It("adds the dependencies", func() {
				err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddReleaseDependencyCallCount()).To(Equal(2))
			})

			Context("when the dependent release ID is zero", func() {
				BeforeEach(func() {
					mdata.Dependencies[1].Release.ID = 0

					dependentRelease := pivnet.Release{
						ID: 9876,
					}

					pivnetClient.GetReleaseReturns(dependentRelease, nil)
				})

				It("obtains the dependent release from the version and product slug", func() {
					err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
					Expect(err).NotTo(HaveOccurred())

					Expect(pivnetClient.AddReleaseDependencyCallCount()).To(Equal(2))
					Expect(pivnetClient.GetReleaseCallCount()).To(Equal(1))
				})

				Context("when obtaining the dependent release returns an error", func() {
					var (
						expectedErr error
					)

					BeforeEach(func() {
						expectedErr = fmt.Errorf("some release error")
						pivnetClient.GetReleaseReturns(pivnet.Release{}, expectedErr)
					})

					It("forwards the error", func() {
						err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when dependent release version is empty", func() {
					BeforeEach(func() {
						mdata.Dependencies[1].Release.Version = ""
					})

					It("returns an error", func() {
						err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err.Error()).To(ContainSubstring("dependencies[1]"))
					})
				})

				Context("when dependent product slug is empty", func() {
					BeforeEach(func() {
						mdata.Dependencies[1].Release.Product = metadata.Product{}
					})

					It("returns an error", func() {
						err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
						Expect(err).To(HaveOccurred())

						Expect(err.Error()).To(ContainSubstring("dependencies[1]"))
					})
				})
			})

			Context("when adding dependency returns an error ", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = fmt.Errorf("boom")
					pivnetClient.AddReleaseDependencyReturns(expectedErr)
				})

				It("returns an error", func() {
					err := releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
