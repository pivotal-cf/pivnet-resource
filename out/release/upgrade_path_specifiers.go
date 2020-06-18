package release

import (
	"fmt"
	pivnet "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type UpgradePathSpecifiersCreator struct {
	logger      logger.Logger
	pivnet      upgradePathSpecifiersCreatorClient
	metadata    metadata.Metadata
	productSlug string
}

func NewUpgradePathSpecifiersCreator(
	logger logger.Logger,
	pivnetClient upgradePathSpecifiersCreatorClient,
	metadata metadata.Metadata,
	productSlug string,
) UpgradePathSpecifiersCreator {
	return UpgradePathSpecifiersCreator{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name UpgradePathSpecifiersCreatorClient . upgradePathSpecifiersCreatorClient
type upgradePathSpecifiersCreatorClient interface {
	CreateUpgradePathSpecifier(productSlug string, releaseID int, specifier string) (pivnet.UpgradePathSpecifier, error)
}

func (creator UpgradePathSpecifiersCreator) CreateUpgradePathSpecifiers(release pivnet.Release) error {
	for _, specifier := range creator.metadata.UpgradePathSpecifiers {
		creator.logger.Info(fmt.Sprintf(
			"Creating upgrade path specifier '%s'",
			specifier.Specifier,
		))
		_, err := creator.pivnet.CreateUpgradePathSpecifier(creator.productSlug, release.ID, specifier.Specifier)
		if err != nil {
			return err
		}
	}

	return nil
}
