package release

import (
	"fmt"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v2/metadata"
)

type ReleaseDependenciesAdder struct {
	logger      logger.Logger
	pivnet      releaseDependenciesAdderClient
	metadata    metadata.Metadata
	productSlug string
}

func NewReleaseDependenciesAdder(
	logger logger.Logger,
	pivnetClient releaseDependenciesAdderClient,
	metadata metadata.Metadata,
	productSlug string,
) ReleaseDependenciesAdder {
	return ReleaseDependenciesAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name ReleaseDependenciesAdderClient . releaseDependenciesAdderClient
type releaseDependenciesAdderClient interface {
	AddReleaseDependency(productSlug string, releaseID int, dependentReleaseID int) error
	GetRelease(productSlug string, releaseVersion string) (pivnet.Release, error)
}

func (rf ReleaseDependenciesAdder) AddReleaseDependencies(release pivnet.Release) error {
	for i, d := range rf.metadata.Dependencies {
		dependentReleaseID := d.Release.ID
		if dependentReleaseID == 0 {
			if d.Release.Version == "" || d.Release.Product.Slug == "" {
				return fmt.Errorf(
					"Either ReleaseID or release version and product slug must be provided for dependencies[%d]",
					i,
				)
			}

			rf.logger.Info(fmt.Sprintf(
				"Looking up dependent release ID for: '%s/%s'",
				d.Release.Product.Slug,
				d.Release.Version,
			))
			r, err := rf.pivnet.GetRelease(d.Release.Product.Slug, d.Release.Version)
			if err != nil {
				return err
			}
			dependentReleaseID = r.ID
		}

		rf.logger.Info(fmt.Sprintf(
			"Adding dependent release with ID: %d",
			dependentReleaseID,
		))
		err := rf.pivnet.AddReleaseDependency(rf.productSlug, release.ID, dependentReleaseID)
		if err != nil {
			return err
		}
	}

	return nil
}
