package filter_test

import (
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/filter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	var (
		fakeLogger logger.Logger

		f *filter.Filter

		releases []pivnet.Release
	)

	BeforeEach(func() {
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		f = filter.NewFilter(fakeLogger)

		releases = []pivnet.Release{
			{
				ID:          1,
				Version:     "version1",
				ReleaseType: pivnet.ReleaseType("foo"),
			},
			{
				ID:          2,
				Version:     "version2",
				ReleaseType: pivnet.ReleaseType("bar"),
			},
			{
				ID:          3,
				Version:     "version3",
				ReleaseType: pivnet.ReleaseType("foo"),
			},
		}
	})

	Describe("ReleasesByReleaseType", func() {
		var (
			releaseType pivnet.ReleaseType
			releases    []pivnet.Release
		)

		BeforeEach(func() {
			releaseType = pivnet.ReleaseType("foo")

			releases = []pivnet.Release{
				{
					ID:          1,
					Version:     "version1",
					ReleaseType: pivnet.ReleaseType("foo"),
				},
				{
					ID:          2,
					Version:     "version2",
					ReleaseType: pivnet.ReleaseType("bar"),
				},
				{
					ID:          3,
					Version:     "version3",
					ReleaseType: pivnet.ReleaseType("foo"),
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
			releases []pivnet.Release
		)

		BeforeEach(func() {
			version = "version2"

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
					Links:        &pivnet.Links{Download: map[string]string{"href": "android-link"}},
				},
				{
					ID:           4,
					AWSObjectKey: "product_files/banana/file-name-2.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "ios-link"}},
				},
			}

			links := f.DownloadLinks(productFiles)
			Expect(links).To(HaveLen(2))
			Expect(links).To(Equal(map[string]string{
				"file-name-1.zip": "android-link",
				"file-name-2.zip": "ios-link",
			}))
		})
	})

	Describe("ProductFileNamesByGlobs", func() {
		var (
			productFiles []pivnet.ProductFile
			globs        []string
		)

		BeforeEach(func() {
			productFiles = []pivnet.ProductFile{
				{
					ID:   1234,
					Name: "name0",
				},
				{
					ID:   2345,
					Name: "name1",
				},
				{
					ID:   3456,
					Name: "name2",
				},
			}

			globs = []string{"*name1*", "*name2*"}
		})

		It("returns the download links that match the glob filters", func() {
			filtered, err := f.ProductFileNamesByGlobs(
				productFiles,
				globs,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(filtered).To(HaveLen(2))
			Expect(filtered).To(Equal([]pivnet.ProductFile{productFiles[1], productFiles[2]}))
		})

		Context("when a bad pattern is passed", func() {
			BeforeEach(func() {
				globs = []string{"["}
			})

			It("returns an error", func() {
				_, err := f.ProductFileNamesByGlobs(
					productFiles,
					globs,
				)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("syntax error in pattern"))
			})
		})

		Describe("Passed a glob that matches no files", func() {
			BeforeEach(func() {
				globs = []string{"*will-not-match*"}
			})

			It("returns empty slice", func() {
				filtered, err := f.ProductFileNamesByGlobs(
					productFiles,
					globs,
				)
				Expect(err).To(HaveOccurred())

				Expect(filtered).To(HaveLen(0))
			})
		})

		Describe("When a glob that matches a file and glob that does not match a file", func() {
			BeforeEach(func() {
				globs = []string{"name1", "does-not-exist.txt"}
			})

			It("returns an error", func() {
				_, err := f.ProductFileNamesByGlobs(
					productFiles,
					globs,
				)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no match for glob: 'does-not-exist.txt'"))
			})
		})
	})
})
