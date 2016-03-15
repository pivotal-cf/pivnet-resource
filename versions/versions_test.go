package versions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("Versions", func() {
	Describe("Since", func() {
		var (
			allVersions []string
			version     string
		)

		BeforeEach(func() {
			allVersions = []string{}
			version = ""
		})

		Context("when the provided version is the newest", func() {
			var (
				allVersions []string
				version     string
			)

			BeforeEach(func() {
				allVersions = []string{"1.2.3#abc", "1.3.2#def"}
				version = "1.2.3#abc"
			})

			It("returns empty array", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(HaveLen(0))
			})
		})

		Context("when provided version is present but not the newest", func() {
			BeforeEach(func() {
				allVersions = []string{"newest version", "middle version", "older version", "last version"}
				version = "older version"
			})

			It("returns new versions", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(Equal([]string{"newest version", "middle version"}))
			})
		})

		Context("When the version is not present", func() {
			BeforeEach(func() {
				allVersions = []string{"1.2.3#abc", "1.3.2#def"}
				version = "1.3.2"
			})

			It("returns the newest version", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(Equal([]string{"1.2.3#abc"}))
			})
		})
	})

	Describe("Reverse", func() {
		It("returns reversed ordered versions because concourse expects them that way", func() {
			versions, _ := versions.Reverse([]string{"v201", "v178", "v120", "v200"})

			Expect(versions).To(Equal([]string{"v200", "v120", "v178", "v201"}))
		})
	})
})
