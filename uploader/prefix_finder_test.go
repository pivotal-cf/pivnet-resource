package uploader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/pivnet-resource/v2/uploader"
	"github.com/pivotal-cf/pivnet-resource/v2/uploader/uploaderfakes"
)

var _ = Describe("PrefixFetcher", func() {
	Context("GetPrefix", func() {
		It("returns the product file prefix", func() {
			productSlug := "product-slug"
			productPrefix := "/my-product/file-prefix"
			fakeS3PrefixFetcher := &uploaderfakes.FakeS3PrefixFetcher{}
			fakeS3PrefixFetcher.S3PrefixForProductSlugReturns(productPrefix, nil)

			prefixFetcher := NewPrefixFetcher(fakeS3PrefixFetcher, productSlug)
			prefix, err := prefixFetcher.GetPrefix()

			Expect(err).NotTo(HaveOccurred())
			Expect(prefix).To(Equal(productPrefix))
		})
	})
})
