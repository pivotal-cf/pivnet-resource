package release_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetadataFetcher", func() {
	Describe("Fetch", func() {
		Context("when skipFileCheck is true", func() {
			It("grabs values from the metadata", func() {
				f := release.NewMetadataFetcher(metadata.Metadata{Release: &metadata.Release{Version: "some-version"}}, true)
				v := f.Fetch("Version", "doesn't matter", "nope")
				Expect(v).To(Equal("some-version"))
			})

			Context("when the key is UserGroupIDs", func() {
				It("returns formatted IDs", func() {
					f := release.NewMetadataFetcher(metadata.Metadata{Release: &metadata.Release{UserGroupIDs: []string{"111", "222"}}}, true)
					v := f.Fetch("UserGroupIDs", "doesn't matter", "nope")
					Expect(v).To(Equal("111,222"))
				})
			})
		})

		Context("when skipFileCheck is false", func() {
			var (
				metadataLocation string
			)

			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				metadataLocation = tempFile.Name()

				ioutil.WriteFile(metadataLocation, []byte("some-data"), 0777)
			})

			AfterEach(func() {
				err := os.Remove(metadataLocation)
				Expect(err).NotTo(HaveOccurred())
			})

			It("grabs values from the file", func() {
				f := release.NewMetadataFetcher(metadata.Metadata{}, false)
				v := f.Fetch("Version", filepath.Dir(metadataLocation), filepath.Base(metadataLocation))
				Expect(v).To(Equal("some-data"))
			})
		})
	})
})
