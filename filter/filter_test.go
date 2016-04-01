package filter_test

import (
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	Describe("Download Links", func() {
		It("returns the download links", func() {
			productFiles := pivnet.ProductFiles{[]pivnet.ProductFile{
				{ID: 3,
					AWSObjectKey: "product_files/banana/file-name-1.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/6/download"}},
				},
				{ID: 4,
					AWSObjectKey: "product_files/banana/file-name-2.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/8/download"}},
				},
			}}

			links := filter.DownloadLinks(productFiles)
			Expect(links).To(HaveLen(2))
			Expect(links).To(Equal(map[string]string{
				"file-name-1.zip": "/products/banana/releases/666/product_files/6/download",
				"file-name-2.zip": "/products/banana/releases/666/product_files/8/download",
			}))
		})
	})

	Describe("Download Links by Glob", func() {
		It("returns the download links that match the glob filters", func() {
			downloadLinks := map[string]string{
				"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				"ios-file.zip":     "/products/banana/releases/666/product_files/8/download",
				"random-file.zip":  "/products/banana/releases/666/product_files/8/download",
			}

			filteredDownloadLinks, err := filter.DownloadLinksByGlob(
				downloadLinks, []string{"*android*", "*ios*"})
			Expect(err).NotTo(HaveOccurred())
			Expect(filteredDownloadLinks).To(HaveLen(2))
			Expect(filteredDownloadLinks).To(Equal(map[string]string{
				"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				"ios-file.zip":     "/products/banana/releases/666/product_files/8/download",
			}))
		})

		Context("when a bad pattern is passed", func() {
			It("returns an error", func() {
				downloadLinks := map[string]string{
					"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				}

				_, err := filter.DownloadLinksByGlob(downloadLinks, []string{"["})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("syntax error in pattern"))
			})
		})

		Describe("Passed a glob that matches no files", func() {
			It("returns an error", func() {
				downloadLinks := map[string]string{
					"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				}

				_, err := filter.DownloadLinksByGlob(downloadLinks, []string{"*ios*"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no files match glob: *ios*"))

			})
		})

		Describe("When a glob that matches a file and glob that does not match a file", func() {
			It("returns an error", func() {
				downloadLinks := map[string]string{
					"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				}

				_, err := filter.DownloadLinksByGlob(downloadLinks, []string{"android-file.zip", "does-not-exist.txt"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no files match glob: does-not-exist.txt"))
			})
		})
	})

	Describe("ReleasesByReleaseType", func() {
		var (
			releases    []pivnet.Release
			releaseType string
		)

		BeforeEach(func() {
			releases = []pivnet.Release{
				{
					ID:          1,
					ReleaseType: "foo",
				},
				{
					ID:          2,
					ReleaseType: "bar",
				},
				{
					ID:          3,
					ReleaseType: "foo",
				},
			}

			releaseType = "foo"
		})

		It("filters releases by release type without error", func() {
			filteredReleases, err := filter.ReleasesByReleaseType(releases, releaseType)

			Expect(err).NotTo(HaveOccurred())

			Expect(filteredReleases).To(HaveLen(2))
			Expect(filteredReleases).To(ContainElement(releases[0]))
			Expect(filteredReleases).To(ContainElement(releases[2]))
		})

		Context("when the input releases are nil", func() {
			It("returns empty slice without error", func() {
				filteredReleases, err := filter.ReleasesByReleaseType(nil, releaseType)

				Expect(err).NotTo(HaveOccurred())

				Expect(filteredReleases).NotTo(BeNil())
				Expect(filteredReleases).To(HaveLen(0))
			})
		})
	})
})
