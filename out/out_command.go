package out

import (
	"fmt"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
)

type OutCommand struct {
	logger                         logger.Logger
	outDir                         string
	sourcesDir                     string
	globClient                     globber
	validation                     validation
	creator                        creator
	finder                         finder
	userGroupsUpdater              userGroupsUpdater
	releaseFileGroupsAdder         releaseFileGroupsAdder
	releaseArtifactReferencesAdder releaseArtifactReferencesAdder
	releaseDependenciesAdder       releaseDependenciesAdder
	dependencySpecifiersCreator    dependencySpecifiersCreator
	releaseUpgradePathsAdder       releaseUpgradePathsAdder
	upgradePathSpecifiersCreator   upgradePathSpecifiersCreator
	finalizer                      finalizer
	uploader                       uploader
	m                              metadata.Metadata
	skipUpload                     bool
	filesOnly                      bool
}

type OutCommandConfig struct {
	Logger                         logger.Logger
	OutDir                         string
	SourcesDir                     string
	GlobClient                     globber
	Validation                     validation
	Creator                        creator
	Finder                         finder
	UserGroupsUpdater              userGroupsUpdater
	ReleaseFileGroupsAdder         releaseFileGroupsAdder
	ReleaseArtifactReferencesAdder releaseArtifactReferencesAdder
	ReleaseDependenciesAdder       releaseDependenciesAdder
	DependencySpecifiersCreator    dependencySpecifiersCreator
	ReleaseUpgradePathsAdder       releaseUpgradePathsAdder
	UpgradePathSpecifiersCreator   upgradePathSpecifiersCreator
	Finalizer                      finalizer
	Uploader                       uploader
	M                              metadata.Metadata
	SkipUpload                     bool
	FilesOnly                      bool
}

func NewOutCommand(config OutCommandConfig) OutCommand {
	return OutCommand{
		logger:                         config.Logger,
		outDir:                         config.OutDir,
		sourcesDir:                     config.SourcesDir,
		globClient:                     config.GlobClient,
		validation:                     config.Validation,
		creator:                        config.Creator,
		finder:                         config.Finder,
		userGroupsUpdater:              config.UserGroupsUpdater,
		releaseFileGroupsAdder:         config.ReleaseFileGroupsAdder,
		releaseArtifactReferencesAdder: config.ReleaseArtifactReferencesAdder,
		releaseDependenciesAdder:       config.ReleaseDependenciesAdder,
		dependencySpecifiersCreator:    config.DependencySpecifiersCreator,
		releaseUpgradePathsAdder:       config.ReleaseUpgradePathsAdder,
		upgradePathSpecifiersCreator:   config.UpgradePathSpecifiersCreator,
		finalizer:                      config.Finalizer,
		uploader:                       config.Uploader,
		m:                              config.M,
		skipUpload:                     config.SkipUpload,
		filesOnly:                      config.FilesOnly,
	}
}

//counterfeiter:generate --fake-name Creator . creator
type creator interface {
	Create() (pivnet.Release, error)
}

//counterfeiter:generate --fake-name Finder . finder
type finder interface {
	Find(i int) (pivnet.Release, error)
}

//counterfeiter:generate --fake-name Uploader . uploader
type uploader interface {
	Upload(release pivnet.Release, exactGlobs []string) error
}

//counterfeiter:generate --fake-name UserGroupsUpdater . userGroupsUpdater
type userGroupsUpdater interface {
	UpdateUserGroups(release pivnet.Release) (pivnet.Release, error)
}

//counterfeiter:generate --fake-name ReleaseFileGroupsAdder . releaseFileGroupsAdder
type releaseFileGroupsAdder interface {
	AddReleaseFileGroups(release pivnet.Release) error
}

//counterfeiter:generate --fake-name ReleaseArtifactReferencesAdder . releaseArtifactReferencesAdder
type releaseArtifactReferencesAdder interface {
	AddReleaseArtifactReferences(release pivnet.Release) error
}

//counterfeiter:generate --fake-name ReleaseDependenciesAdder . releaseDependenciesAdder
type releaseDependenciesAdder interface {
	AddReleaseDependencies(release pivnet.Release) error
}

//counterfeiter:generate --fake-name DependencySpecifiersCreator . dependencySpecifiersCreator
type dependencySpecifiersCreator interface {
	CreateDependencySpecifiers(release pivnet.Release) error
}

//counterfeiter:generate --fake-name ReleaseUpgradePathsAdder . releaseUpgradePathsAdder
type releaseUpgradePathsAdder interface {
	AddReleaseUpgradePaths(release pivnet.Release) error
}

//counterfeiter:generate --fake-name UpgradePathSpecifiersCreator . upgradePathSpecifiersCreator
type upgradePathSpecifiersCreator interface {
	CreateUpgradePathSpecifiers(release pivnet.Release) error
}

//counterfeiter:generate --fake-name Finalizer . finalizer
type finalizer interface {
	Finalize(productSlug string, releaseVersion string) (concourse.OutResponse, error)
}

//counterfeiter:generate --fake-name Validation . validation
type validation interface {
	Validate() error
}

//counterfeiter:generate --fake-name Globber . globber
type globber interface {
	ExactGlobs() ([]string, error)
}

func (c OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	var out concourse.OutResponse
	if c.outDir == "" {
		return concourse.OutResponse{}, fmt.Errorf("out dir must be provided")
	}

	err := c.validation.Validate()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	exactGlobs, err := c.globClient.ExactGlobs()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	var missingFiles []string
	for _, f := range c.m.ProductFiles {
		var foundFile bool
		for _, glob := range exactGlobs {
			if glob == f.File {
				foundFile = true
				continue
			}
		}

		if !foundFile {
			missingFiles = append(missingFiles, f.File)
			foundFile = false
		}
	}

	if len(missingFiles) > 0 {
		return concourse.OutResponse{},
			fmt.Errorf(
				"product files were provided in metadata that match no globs: %v",
				missingFiles,
			)
	}

	var pivnetRelease pivnet.Release
	if !c.filesOnly {
		pivnetRelease, err = c.creator.Create()
		if err != nil {
			return concourse.OutResponse{}, err
		}
	} else {
		pivnetRelease, err = c.finder.Find(c.m.ExistingRelease.ID)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}

	if c.skipUpload {
		c.logger.Info(
			"file glob not provided - skipping upload to s3")
	} else {
		err = c.uploader.Upload(pivnetRelease, exactGlobs)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}

	if !c.filesOnly {
		err = c.releaseFileGroupsAdder.AddReleaseFileGroups(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		err = c.releaseArtifactReferencesAdder.AddReleaseArtifactReferences(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		err = c.releaseUpgradePathsAdder.AddReleaseUpgradePaths(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		err = c.releaseDependenciesAdder.AddReleaseDependencies(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		err = c.upgradePathSpecifiersCreator.CreateUpgradePathSpecifiers(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		err = c.dependencySpecifiersCreator.CreateDependencySpecifiers(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		pivnetRelease, err = c.userGroupsUpdater.UpdateUserGroups(pivnetRelease)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}
	out, err = c.finalizer.Finalize(input.Source.ProductSlug, pivnetRelease.Version)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Info("Put complete")

	return out, nil
}
