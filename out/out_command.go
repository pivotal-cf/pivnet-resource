package out

import (
	"fmt"
	"log"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

type OutCommand struct {
	skipFileCheck bool
	logger        *log.Logger
	outDir        string
	sourcesDir    string
	globClient    globber
	validation    validation
	creator       creator
	finalizer     finalizer
	uploader      uploader
	m             metadata.Metadata
}

type OutCommandConfig struct {
	SkipFileCheck bool
	Logger        *log.Logger
	OutDir        string
	SourcesDir    string
	GlobClient    globber
	Validation    validation
	Creator       creator
	Finalizer     finalizer
	Uploader      uploader
	M             metadata.Metadata
}

func NewOutCommand(config OutCommandConfig) OutCommand {
	return OutCommand{
		skipFileCheck: config.SkipFileCheck,
		logger:        config.Logger,
		outDir:        config.OutDir,
		sourcesDir:    config.SourcesDir,
		globClient:    config.GlobClient,
		validation:    config.Validation,
		creator:       config.Creator,
		finalizer:     config.Finalizer,
		uploader:      config.Uploader,
		m:             config.M,
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

//go:generate counterfeiter --fake-name Finalizer . finalizer
type finalizer interface {
	Finalize(release pivnet.Release) (concourse.OutResponse, error)
}

//go:generate counterfeiter --fake-name Validation . validation
type validation interface {
	Validate(skipFileCheck bool) error
}

//go:generate counterfeiter --fake-name Globber . globber
type globber interface {
	ExactGlobs() ([]string, error)
}

func (c OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	if c.outDir == "" {
		return concourse.OutResponse{}, fmt.Errorf("out dir must be provided")
	}

	if c.m.Release != nil {
		c.logger.Println("metadata release parsed")
	}

	warnIfDeprecatedFilesFound(input.Params, c.logger)

	c.logger.Println("validating metadata")
	err := c.validation.Validate(c.skipFileCheck)
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
				"product_files were provided in metadata that match no globs: %v",
				missingFiles,
			)
	}

	pivnetRelease, err := c.creator.Create()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	err = c.uploader.Upload(pivnetRelease, exactGlobs)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	out, err := c.finalizer.Finalize(pivnetRelease)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	return out, nil
}

func warnIfDeprecatedFilesFound(params concourse.OutParams, logger *log.Logger) {
	files := map[string]string{
		"version_file":        params.VersionFile,
		"eula_slug_file":      params.EULASlugFile,
		"release_date_file":   params.ReleaseDateFile,
		"description_file":    params.DescriptionFile,
		"release_type_file":   params.ReleaseTypeFile,
		"user_group_ids_file": params.UserGroupIDsFile,
		"availability_file":   params.AvailabilityFile,
		"release_notes_file":  params.ReleaseNotesURLFile,
	}
	for key, value := range files {
		if value == "" {
			continue
		}

		logger.Printf("\x1b[31mDEPRECATION WARNING: %q is deprecated and will be removed in a future release\x1b[0m", key)
	}
}
