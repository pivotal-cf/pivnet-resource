package pivnet_filter_test

import (
	gp "github.com/pivotal-cf-experimental/go-pivnet"
	gp_filter "github.com/pivotal-cf-experimental/pivnet-resource/gp_filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	var (
		f gp_filter.Filter

		releases []pivnet.Release
	)

	BeforeEach(func() {
		f = gp_filter.NewFilter()

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
			}
		})

		It("filters releases by release type without error", func() {
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
	})
})
