package release

import "github.com/pivotal-cf/go-pivnet/v7"

type ReleaseFinder struct {
	pivnet      finderClient
	productSlug string
}

//counterfeiter:generate --fake-name=FinderClient . finderClient
type finderClient interface {
	FindRelease(productSlug string, releaseID int) (pivnet.Release, error)
}

func NewReleaseFinder(
	pivnet finderClient,
	productSlug string,
) ReleaseFinder {
	return ReleaseFinder{
		pivnet:      pivnet,
		productSlug: productSlug,
	}
}

func (rf ReleaseFinder) Find(i int) (pivnet.Release, error) {
	return rf.pivnet.FindRelease(rf.productSlug, i)
}
