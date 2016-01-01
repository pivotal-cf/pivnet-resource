package filter_test

import (
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	Describe("Download Links", func() {
		It("returns the download links", func() {
			productFiles := pivnet.ProductFiles{[]pivnet.ProductFile{
				{ID: 3,
					AWSObjectKey: "product_files/banana/file-name-1.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/6/download"}},
				},
				{ID: 4,
					AWSObjectKey: "product_files/banana/file-name-2.zip",
					Links:        &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/8/download"}},
				},
			}}

			links := filter.DownloadLinks(productFiles)
			Expect(links).To(HaveLen(2))
			Expect(links).To(Equal(map[string]string{
				"file-name-1.zip": "/products/banana/releases/666/product_files/6/download",
				"file-name-2.zip": "/products/banana/releases/666/product_files/8/download",
			}))
		})
	})
})
