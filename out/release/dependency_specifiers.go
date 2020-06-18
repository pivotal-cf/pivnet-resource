package release

import (
	"fmt"

	pivnet "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type DependencySpecifiersCreator struct {
	logger      logger.Logger
	pivnet      dependencySpecifiersCreatorClient
	metadata    metadata.Metadata
	productSlug string
}

func NewDependencySpecifiersCreator(
	logger logger.Logger,
	pivnetClient dependencySpecifiersCreatorClient,
	metadata metadata.Metadata,
	productSlug string,
) DependencySpecifiersCreator {
	return DependencySpecifiersCreator{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name DependencySpecifiersCreatorClient . dependencySpecifiersCreatorClient
type dependencySpecifiersCreatorClient interface {
	CreateDependencySpecifier(productSlug string, releaseID int, dependentProductSlug string, specifier string) (pivnet.DependencySpecifier, error)
}

func (rf DependencySpecifiersCreator) CreateDependencySpecifiers(release pivnet.Release) error {
	for _, d := range rf.metadata.DependencySpecifiers {
		rf.logger.Info(fmt.Sprintf(
			"Creating dependency specifier for: '%s/%s'",
			d.ProductSlug,
			d.Specifier,
		))
		_, err := rf.pivnet.CreateDependencySpecifier(rf.productSlug, release.ID, d.ProductSlug, d.Specifier)
		if err != nil {
			return err
		}
	}

	return nil
}
