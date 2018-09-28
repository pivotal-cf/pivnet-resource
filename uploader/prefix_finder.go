package uploader

//go:generate counterfeiter --fake-name FakeS3PrefixFetcher . S3PrefixFetcher
type S3PrefixFetcher interface {
	S3PrefixForProductSlug(productSlug string) (string, error)
}

type PrefixFetcher struct {
	productSlug     string
	s3PrefixFetcher S3PrefixFetcher
}

func NewPrefixFetcher(fetcher S3PrefixFetcher, productSlug string) PrefixFetcher {
	return PrefixFetcher{
		productSlug: productSlug,
		s3PrefixFetcher: fetcher,
	}
}

func (pf *PrefixFetcher) GetPrefix() (string, error) {
	return pf.s3PrefixFetcher.S3PrefixForProductSlug(pf.productSlug)
}
