package sorter_test

import (
	"github.com/pivotal-cf-experimental/pivnet-resource/sorter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorter", func() {
	var (
		s sorter.Sorter
	)

	BeforeEach(func() {
		s = sorter.NewSorter()
	})

	Describe("SortBySemver", func() {
		It("sorts, highest first", func() {
			input := []string{"1.0.0", "2.4.1", "2.0.0", "2.4.1-edge.12", "2.4.1-edge.11"}

			returned, err := s.SortBySemver(input)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal([]string{"2.4.1", "2.4.1-edge.12", "2.4.1-edge.11", "2.0.0", "1.0.0"}))
		})

		Context("when parsing the semver fails", func() {
			It("returns error", func() {
				input := []string{"1.0.0", "2.4.1", "not-semver"}

				_, err := s.SortBySemver(input)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the versions have fewer than 3 components", func() {
			It("treats the missing components as zero", func() {
				input := []string{"1", "2.4.1", "2.1"}

				returned, err := s.SortBySemver(input)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal([]string{"2.4.1", "2.1.0", "1.0.0"}))
			})
		})

		Context("when the versions have more than 3 components", func() {
			It("returns error", func() {
				input := []string{"1.0.0.1", "2.4.1", "2.1"}

				_, err := s.SortBySemver(input)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
