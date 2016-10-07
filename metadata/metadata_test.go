package metadata_test

import (
	"fmt"

	"github.com/pivotal-cf/pivnet-resource/metadata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metadata", func() {
	Describe("Validate", func() {
		var data metadata.Metadata
		BeforeEach(func() {
			data = metadata.Metadata{
				Release: &metadata.Release{
					Version:     "1.0.0",
					ReleaseType: "All In One",
					EULASlug:    "some-other-eula",
				},
				ProductFiles: []metadata.ProductFile{
					{File: "hello.txt", Description: "available"},
				},
			}
		})

		It("returns an error when product files are missing", func() {
			data.ProductFiles[0].File = ""
			Expect(data.Validate()).To(MatchError("empty value for file"))
		})

		It("returns an error when eula slug is missing", func() {
			data.Release.EULASlug = ""
			Expect(data.Validate()).To(MatchError(fmt.Sprintf("missing required value %q", "eula_slug")))
		})

		It("returns an error when version is missing", func() {
			data.Release.Version = ""
			Expect(data.Validate()).To(MatchError(fmt.Sprintf("missing required value %q", "version")))
		})

		It("returns an error when release type is missing", func() {
			data.Release.ReleaseType = ""
			Expect(data.Validate()).To(MatchError(fmt.Sprintf("missing required value %q", "release_type")))
		})

		Context("when no top-level release key is provided", func() {
			It("does not perform any validations", func() {
				data = metadata.Metadata{
					ProductFiles: []metadata.ProductFile{
						{File: "hello.txt", Description: "available"},
					},
				}
				Expect(data.Validate()).NotTo(HaveOccurred())
			})
		})

		Context("when dependencies exist with id 0", func() {
			BeforeEach(func() {
				data.Dependencies = []metadata.Dependency{
					{
						Release: metadata.DependentRelease{
							ID:      0,
							Version: "abcd",
							Product: metadata.Product{
								Slug: "some-product",
							},
						},
					},
				}
			})

			It("returns without error", func() {
				Expect(data.Validate()).NotTo(HaveOccurred())
			})

			Context("when release version is empty", func() {
				BeforeEach(func() {
					data.Dependencies[0].Release.Version = ""
				})

				It("returns an error", func() {
					err := data.Validate()
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(MatchRegexp(".*dependency\\[0\\]"))
				})
			})

			Context("when produc slug is empty", func() {
				BeforeEach(func() {
					data.Dependencies[0].Release.Product.Slug = ""
				})

				It("returns an error", func() {
					err := data.Validate()
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(MatchRegexp(".*dependency\\[0\\]"))
				})
			})
		})
	})
})
