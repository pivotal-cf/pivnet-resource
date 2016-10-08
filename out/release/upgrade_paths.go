package release

import (
	"fmt"
	"log"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseUpgradePathsAdder struct {
	logger      *log.Logger
	pivnet      releaseUpgradePathsAdderClient
	metadata    metadata.Metadata
	productSlug string
}

func NewReleaseUpgradePathsAdder(
	logger *log.Logger,
	pivnetClient releaseUpgradePathsAdderClient,
	metadata metadata.Metadata,
	productSlug string,
) ReleaseUpgradePathsAdder {
	return ReleaseUpgradePathsAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name ReleaseUpgradePathsAdderClient . releaseUpgradePathsAdderClient
type releaseUpgradePathsAdderClient interface {
	AddReleaseUpgradePath(productSlug string, releaseID int, previousReleaseID int) error
	GetRelease(productSlug string, releaseVersion string) (pivnet.Release, error)
}

func (rf ReleaseUpgradePathsAdder) AddReleaseUpgradePaths(release pivnet.Release) error {
	for i, u := range rf.metadata.UpgradePaths {
		previousReleaseID := u.ID
		if previousReleaseID == 0 {
			if u.Version == "" {
				return fmt.Errorf(
					"Either id or version must be provided for upgrade_paths[%d]",
					i,
				)
			}

			rf.logger.Println(fmt.Sprintf(
				"Looking up previous release ID for: '%s/%s'",
				rf.productSlug,
				u.Version,
			))
			r, err := rf.pivnet.GetRelease(rf.productSlug, u.Version)
			if err != nil {
				return err
			}
			previousReleaseID = r.ID
		}

		rf.logger.Println(fmt.Sprintf(
			"Adding upgrade path for release with ID: %d",
			previousReleaseID,
		))
		err := rf.pivnet.AddReleaseUpgradePath(rf.productSlug, release.ID, previousReleaseID)
		if err != nil {
			return err
		}
	}

	return nil
}
