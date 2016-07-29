package release_test

import (
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetadataFetcher", func() {
	Describe("Fetch", func() {
		It("grabs values from the metadata", func() {
			f := release.NewMetadataFetcher(metadata.Metadata{Release: &metadata.Release{Version: "some-version"}})
			v := f.Fetch("Version")
			Expect(v).To(Equal("some-version"))
		})

		Context("when the key is UserGroupIDs", func() {
			It("returns formatted IDs", func() {
				f := release.NewMetadataFetcher(metadata.Metadata{Release: &metadata.Release{UserGroupIDs: []string{"111", "222"}}})
				v := f.Fetch("UserGroupIDs")
				Expect(v).To(Equal("111,222"))
			})
		})

		Context("when no release is passed", func() {
			It("returns empty string", func() {
				f := release.NewMetadataFetcher(metadata.Metadata{})
				v := f.Fetch("No")
				Expect(v).To(Equal(""))
			})
		})
	})
})
