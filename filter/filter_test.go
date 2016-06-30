package filter_test

import (
	"github.com/pivotal-cf-experimental/go-pivnet"
	gp "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	var (
		f filter.Filter

		releases []pivnet.Release
	)

	BeforeEach(func() {
		f = filter.NewFilter()

		releases = []pivnet.Release{
			{
				ID:          1,
				Version:     "version1",
				ReleaseType: "foo",
			},
			{
				ID:          2,
				Version:     "version2",
				ReleaseType: "bar",
			},
			{
				ID:          3,
				Version:     "version3",
				ReleaseType: "foo",
			},
		}
	})

	Describe("ReleasesByReleaseType", func() {
		var (
			releaseType string
			releases    []gp.Release
		)

		BeforeEach(func() {
			releaseType = "foo"

			releases = []gp.Release{
				{
					ID:          1,
					Version:     "version1",
					ReleaseType: "foo",
				},
				{
					ID:          2,
					Version:     "version2",
					ReleaseType: "bar",
				},
				{
					ID:          3,
					Version:     "version3",
					ReleaseType: "foo",
				},
			}
		})

		It("filters releases by release type without error", func() {
			filteredReleases, err := f.ReleasesByReleaseType(releases, releaseType)

			Expect(err).NotTo(HaveOccurred())

			Expect(filteredReleases).To(HaveLen(2))
			Expect(filteredReleases).To(ContainElement(releases[0]))
			Expect(filteredReleases).To(ContainElement(releases[2]))
		})

		Context("when the input releases are nil", func() {
			BeforeEach(func() {
				releases = nil
			})

			It("returns empty slice without error", func() {
				filteredReleases, err := f.ReleasesByReleaseType(releases, releaseType)

				Expect(err).NotTo(HaveOccurred())

				Expect(filteredReleases).NotTo(BeNil())
				Expect(filteredReleases).To(HaveLen(0))
			})
		})
	})

	Describe("ReleasesByVersion", func() {
		var (
			version  string
			releases []gp.Release
		)

		BeforeEach(func() {
			version = "version2"

			releases = []gp.Release{
				{
					ID:          1,
					Version:     "version1",
					ReleaseType: "foo",
				},
				{
					ID:          2,
					Version:     "version2",
					ReleaseType: "bar",
				},
				{
					ID:          3,
					Version:     "version3",
					ReleaseType: "foo",
				},
				{
					ID:          4,
					Version:     "version3.2",
					ReleaseType: "foo-minor",
				},
				{
					ID:          5,
					Version:     "version3.1.2",
					ReleaseType: "foo-patch",
				},
			}
		})

		It("filters releases by version without error", func() {
			filteredReleases, err := f.ReleasesByVersion(releases, version)

			Expect(err).NotTo(HaveOccurred())

			Expect(filteredReleases).To(HaveLen(1))
			Expect(filteredReleases).To(ContainElement(releases[1]))
		})

		Context("when the input releases are nil", func() {
			BeforeEach(func() {
				releases = nil
			})

			It("returns empty slice without error", func() {
				filteredReleases, err := f.ReleasesByVersion(releases, version)

				Expect(err).NotTo(HaveOccurred())

				Expect(filteredReleases).NotTo(BeNil())
				Expect(filteredReleases).To(HaveLen(0))
			})
		})

		Describe("Matching on regex", func() {
			Context("when the regex matches one release versions", func() {
				BeforeEach(func() {
					version = `version3\.1\..*`
				})

				It("returns all releases that match the regex without error", func() {
					filteredReleases, err := f.ReleasesByVersion(releases, version)

					Expect(err).NotTo(HaveOccurred())

					Expect(filteredReleases).To(HaveLen(1))
					Expect(filteredReleases).To(ContainElement(releases[4]))
				})
			})

			Context("when the regex matches multiple release versions", func() {
				BeforeEach(func() {
					version = `version3\..*`
				})

				It("returns all releases that match the regex without error", func() {
					filteredReleases, err := f.ReleasesByVersion(releases, version)

					Expect(err).NotTo(HaveOccurred())

					Expect(filteredReleases).To(HaveLen(2))
					Expect(filteredReleases).To(ContainElement(releases[3]))
					Expect(filteredReleases).To(ContainElement(releases[4]))
				})
			})

			Context("when the regex is invalid", func() {
				BeforeEach(func() {
					version = "some(invalid^regex"
				})

				It("returns error", func() {
					_, err := f.ReleasesByVersion(releases, version)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Download Links", func() {
		It("returns the download links", func() {
			productFiles := []pivnet.ProductFile{
				{
					ID:           3,
					AWSObjectKey: "product_files/banana/file-name-1.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/6/download"}},
				},
				{
					ID:           4,
					AWSObjectKey: "product_files/banana/file-name-2.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/8/download"}},
				},
			}

			links := f.DownloadLinks(productFiles)
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

			filteredDownloadLinks, err := f.DownloadLinksByGlob(
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

				_, err := f.DownloadLinksByGlob(downloadLinks, []string{"["})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("syntax error in pattern"))
			})
		})

		Describe("Passed a glob that matches no files", func() {
			It("returns an error", func() {
				downloadLinks := map[string]string{
					"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				}

				_, err := f.DownloadLinksByGlob(downloadLinks, []string{"*ios*"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no files match glob: *ios*"))

			})
		})

		Describe("When a glob that matches a file and glob that does not match a file", func() {
			It("returns an error", func() {
				downloadLinks := map[string]string{
					"android-file.zip": "/products/banana/releases/666/product_files/6/download",
				}

				_, err := f.DownloadLinksByGlob(downloadLinks, []string{"android-file.zip", "does-not-exist.txt"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no files match glob: does-not-exist.txt"))
			})
		})
	})
})
