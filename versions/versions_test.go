package versions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("Versions", func() {
	Describe("Since", func() {
		It("returns new versions", func() {
			allVersions := []string{"newest version", "newish version", "oldest version"}
			versions, _ := versions.Since(allVersions, "newish version")

			Expect(versions).To(Equal([]string{"newest version"}))
		})

		Context("when the versions are not ordered", func() {
			var allVersions []string

			BeforeEach(func() {
				allVersions = []string{"aaa", "ddd", "eee", "bbb", "fff", "ccc"}
			})

			It("returns new versions", func() {
				versions, _ := versions.Since(allVersions, "eee")

				Expect(versions).To(Equal([]string{"aaa", "ddd"}))
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
