package out

import (
	"fmt"
	"log"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type OutCommand struct {
	logger                   *log.Logger
	outDir                   string
	sourcesDir               string
	globClient               globber
	validation               validation
	creator                  creator
	userGroupsUpdater        userGroupsUpdater
	releaseDependenciesAdder releaseDependenciesAdder
	releaseUpgradePathsAdder releaseUpgradePathsAdder
	finalizer                finalizer
	uploader                 uploader
	m                        metadata.Metadata
	skipUpload               bool
}

type OutCommandConfig struct {
	Logger                   *log.Logger
	OutDir                   string
	SourcesDir               string
	GlobClient               globber
	Validation               validation
	Creator                  creator
	UserGroupsUpdater        userGroupsUpdater
	ReleaseDependenciesAdder releaseDependenciesAdder
	ReleaseUpgradePathsAdder releaseUpgradePathsAdder
	Finalizer                finalizer
	Uploader                 uploader
	M                        metadata.Metadata
	SkipUpload               bool
}

func NewOutCommand(config OutCommandConfig) OutCommand {
	return OutCommand{
		logger:                   config.Logger,
		outDir:                   config.OutDir,
		sourcesDir:               config.SourcesDir,
		globClient:               config.GlobClient,
		validation:               config.Validation,
		creator:                  config.Creator,
		userGroupsUpdater:        config.UserGroupsUpdater,
		releaseDependenciesAdder: config.ReleaseDependenciesAdder,
		releaseUpgradePathsAdder: config.ReleaseUpgradePathsAdder,
		finalizer:                config.Finalizer,
		uploader:                 config.Uploader,
		m:                        config.M,
		skipUpload:               config.SkipUpload,
	}
}

//go:generate counterfeiter --fake-name Creator . creator
type creator interface {
	Create() (pivnet.Release, error)
}

//go:generate counterfeiter --fake-name Uploader . uploader
type uploader interface {
	Upload(release pivnet.Release, exactGlobs []string) error
}

//go:generate counterfeiter --fake-name UserGroupsUpdater . userGroupsUpdater
type userGroupsUpdater interface {
	UpdateUserGroups(release pivnet.Release) (pivnet.Release, error)
}

//go:generate counterfeiter --fake-name ReleaseDependenciesAdder . releaseDependenciesAdder
type releaseDependenciesAdder interface {
	AddReleaseDependencies(release pivnet.Release) error
}

//go:generate counterfeiter --fake-name ReleaseUpgradePathsAdder . releaseUpgradePathsAdder
type releaseUpgradePathsAdder interface {
	AddReleaseUpgradePaths(release pivnet.Release) error
}

//go:generate counterfeiter --fake-name Finalizer . finalizer
type finalizer interface {
	Finalize(productSlug string, releaseVersion string) (concourse.OutResponse, error)
}

//go:generate counterfeiter --fake-name Validation . validation
type validation interface {
	Validate() error
}

//go:generate counterfeiter --fake-name Globber . globber
type globber interface {
	ExactGlobs() ([]string, error)
}

func (c OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
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

	pivnetRelease, err := c.creator.Create()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	if c.skipUpload {
		c.logger.Println(
			"file glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		err = c.uploader.Upload(pivnetRelease, exactGlobs)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}

	pivnetRelease, err = c.userGroupsUpdater.UpdateUserGroups(pivnetRelease)
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

	out, err := c.finalizer.Finalize(input.Source.ProductSlug, pivnetRelease.Version)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Println("Put complete")

	return out, nil
}
