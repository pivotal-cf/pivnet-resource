package release

import (
	"fmt"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseUpgradePathsAdder struct {
	logger      logger.Logger
	pivnet      releaseUpgradePathsAdderClient
	metadata    metadata.Metadata
	productSlug string
	filter      filter
}

func NewReleaseUpgradePathsAdder(
	logger logger.Logger,
	pivnetClient releaseUpgradePathsAdderClient,
	metadata metadata.Metadata,
	productSlug string,
	filter filter,
) ReleaseUpgradePathsAdder {
	return ReleaseUpgradePathsAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
		filter:      filter,
	}
}

//go:generate counterfeiter --fake-name ReleaseUpgradePathsAdderClient . releaseUpgradePathsAdderClient
type releaseUpgradePathsAdderClient interface {
	AddReleaseUpgradePath(productSlug string, releaseID int, previousReleaseID int) error
	ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error)
}

//go:generate counterfeiter --fake-name FakeFilter . filter
type filter interface {
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
}

func (rf ReleaseUpgradePathsAdder) AddReleaseUpgradePaths(release pivnet.Release) error {
	allReleases, err := rf.pivnet.ReleasesForProductSlug(rf.productSlug)
	if err != nil {
		return err
	}

	upgradeFromReleases := map[pivnet.Release]interface{}{}

	for i, u := range rf.metadata.UpgradePaths {
		if u.ID == 0 && u.Version == "" {
			return fmt.Errorf(
				"Either id or version must be provided for upgrade_paths[%d]",
				i,
			)
		}

		if u.ID == 0 {
			matchingReleases, err := rf.filter.ReleasesByVersion(allReleases, u.Version)
			if err != nil {
				return err
			}

			if len(matchingReleases) == 0 {
				return fmt.Errorf("No releases found for version: '%s'", u.Version)
			}

			for _, r := range matchingReleases {
				upgradeFromReleases[r] = nil
			}
		} else {
			r, err := filterReleasesForID(allReleases, u.ID)
			if err != nil {
				return err
			}

			upgradeFromReleases[r] = nil
		}
	}

	for r, _ := range upgradeFromReleases {
		rf.logger.Info(fmt.Sprintf(
			"Adding upgrade path: '%s'",
			r.Version,
		))

		if r.ID == release.ID {
			rf.logger.Info(fmt.Sprintf("skipping release: %s", r.Version))
			continue
		}

		err := rf.pivnet.AddReleaseUpgradePath(rf.productSlug, release.ID, r.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func filterReleasesForID(releases []pivnet.Release, id int) (pivnet.Release, error) {
	for _, r := range releases {
		if r.ID == id {
			return r, nil
		}
	}

	return pivnet.Release{}, fmt.Errorf("No releases found for id: '%d'", id)
}
