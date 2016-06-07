package out

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/globs"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
)

type OutCommand struct {
	logger         logger.Logger
	outDir         string
	sourcesDir     string
	screenWriter   *log.Logger
	pivnetClient   pivnet.Client
	uploaderClient uploader.Client
	globClient     globs.Globber
	validation     validation
}

type OutCommandConfig struct {
	Logger         logger.Logger
	OutDir         string
	SourcesDir     string
	ScreenWriter   *log.Logger
	PivnetClient   pivnet.Client
	UploaderClient uploader.Client
	GlobClient     globs.Globber
	Validation     validation
}

func NewOutCommand(config OutCommandConfig) *OutCommand {
	return &OutCommand{
		logger:         config.Logger,
		outDir:         config.OutDir,
		sourcesDir:     config.SourcesDir,
		screenWriter:   config.ScreenWriter,
		pivnetClient:   config.PivnetClient,
		uploaderClient: config.UploaderClient,
		globClient:     config.GlobClient,
		validation:     config.Validation,
	}
}

type validation interface {
	Validate(skipFileCheck bool) error
}

func (c *OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	if c.outDir == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "out dir")
	}

	var m metadata.Metadata
	var skipFileCheck bool
	if input.Params.MetadataFile != "" {
		metadataFilepath := filepath.Join(c.sourcesDir, input.Params.MetadataFile)
		metadataBytes, err := ioutil.ReadFile(metadataFilepath)
		if err != nil {
			return concourse.OutResponse{}, fmt.Errorf("metadata_file could not be read: %s", err.Error())
		}

		err = yaml.Unmarshal(metadataBytes, &m)
		if err != nil {
			return concourse.OutResponse{}, fmt.Errorf("metadata_file could not be parsed: %s", err.Error())
		}

		err = m.Validate()
		if err != nil {
			return concourse.OutResponse{}, fmt.Errorf("metadata_file is invalid: %s", err.Error())
		}

		skipFileCheck = true
	}

	c.logger.Debugf("metadata product_files parsed; contents: %+v\n", m.ProductFiles)

	if m.Release != nil {
		c.logger.Debugf("metadata release parsed; contents: %+v\n", *m.Release)
	}

	warnIfDeprecatedFilesFound(input.Params, c.logger, c.screenWriter)

	err := c.validation.Validate(skipFileCheck)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Debugf("Received input: %+v\n", input)

	exactGlobs, err := c.globClient.ExactGlobs()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	var missingFiles []string
	for _, f := range m.ProductFiles {
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
			fmt.Errorf("product_files were provided in metadata that match no globs: %v", missingFiles)
	}

	metadataFetcher := release.NewMetadataFetcher(m, skipFileCheck)

	releaseCreator := release.NewReleaseCreator(c.pivnetClient, metadataFetcher, c.logger, m, skipFileCheck, input.Params, c.sourcesDir, input.Source.ProductSlug)
	pivnetRelease, err := releaseCreator.Create()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	md5summer := md5sum.NewFileSummer()

	releaseUploader := release.NewReleaseUploader(c.uploaderClient, c.pivnetClient, c.logger, md5summer, m, skipUpload, c.sourcesDir, input.Source.ProductSlug)
	err = releaseUploader.Upload(pivnetRelease, exactGlobs)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	releaseFinalizer := release.NewFinalizer(c.pivnetClient, metadataFetcher, input.Params, c.sourcesDir, input.Source.ProductSlug)

	out := releaseFinalizer.Finalize(pivnetRelease)

	return out, nil
}

func warnIfDeprecatedFilesFound(
	params concourse.OutParams,
	logger logger.Logger,
	screenWriter *log.Logger,
) {
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

		logger.Debugf("\x1b[31mDEPRECATION WARNING: %q is deprecated and will be removed in a future release\x1b[0m\n", key)

		if screenWriter != nil {
			screenWriter.Printf("\x1b[31mDEPRECATION WARNING: %q is deprecated and will be removed in a future release\x1b[0m", key)
		}
	}
}
