package versions_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/versions"
	"github.com/pivotal-cf/pivnet-resource/versions/versionsfakes"
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
			versions, err := versions.Reverse([]string{"v201", "v178", "v120", "v200"})

			Expect(err).NotTo(HaveOccurred())
			Expect(versions).To(Equal([]string{"v200", "v120", "v178", "v201"}))
		})
	})

	Describe("SplitIntoVersionAndETag", func() {
		var (
			input string
		)

		BeforeEach(func() {
			input = "some.version#my-etag"
		})

		It("splits without error", func() {
			version, etag, err := versions.SplitIntoVersionAndETag(input)

			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("some.version"))
			Expect(etag).To(Equal("my-etag"))
		})

		Context("when the input does not contain enough delimiters", func() {
			BeforeEach(func() {
				input = "some.version"
			})

			It("returns error", func() {
				_, _, err := versions.SplitIntoVersionAndETag(input)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the input contains too many delimiters", func() {
			BeforeEach(func() {
				input = "some.version#etag-1#-etag-2"
			})

			It("returns error", func() {
				_, _, err := versions.SplitIntoVersionAndETag(input)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CombineVersionAndETag", func() {
		var (
			version string
			etag    string
		)

		BeforeEach(func() {
			version = "some.version"
			etag = "my-etag"
		})

		It("combines without error", func() {
			versionWithETag, err := versions.CombineVersionAndETag(version, etag)

			Expect(err).NotTo(HaveOccurred())
			Expect(versionWithETag).To(Equal("some.version#my-etag"))
		})

		Context("when the etag is empty", func() {
			BeforeEach(func() {
				etag = ""
			})

			It("does not include the #", func() {
				versionWithETag, err := versions.CombineVersionAndETag(version, etag)

				Expect(err).NotTo(HaveOccurred())
				Expect(versionWithETag).To(Equal("some.version"))
			})
		})
	})

	Describe("ProductVersions", func() {
		var (
			productSlug        string
			releases           []pivnet.Release
			fakeExtendedClient *versionsfakes.FakeExtendedClient
			releaseETagErr     error

			etags []string
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			fakeExtendedClient = &versionsfakes.FakeExtendedClient{}
			releases = []pivnet.Release{
				{
					ID:      1234,
					Version: "version-0",
				},
				{
					ID:      2345,
					Version: "version-1",
				},
			}

			etags = []string{
				"etag-0",
				"etag-1",
			}

			releaseETagErr = nil
		})

		JustBeforeEach(func() {
			fakeExtendedClient.ReleaseETagStub = func(_ string, releaseID int) (string, error) {
				switch releaseID {
				case releases[0].ID:
					return etags[0], releaseETagErr
				case releases[1].ID:
					return etags[1], releaseETagErr
				default:
					panic(fmt.Sprintf("Unexpected releaseID: %d", releaseID))
				}
			}
		})

		It("returns fetched etag to product version", func() {
			versions, err := versions.ProductVersions(
				fakeExtendedClient,
				productSlug,
				releases,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(versions)).To(Equal(len(releases)))
			Expect(versions[0]).To(Equal("version-0#etag-0"))
			Expect(versions[1]).To(Equal("version-1#etag-1"))
		})

		Context("when fetching etag returns an error", func() {
			BeforeEach(func() {
				releaseETagErr = errors.New("some etag error")
			})

			It("returns the error", func() {
				_, err := versions.ProductVersions(
					fakeExtendedClient,
					productSlug,
					releases,
				)
				Expect(err).To(Equal(releaseETagErr))
			})
		})
	})
})
