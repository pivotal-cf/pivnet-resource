package out

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/s3"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
)

const (
	defaultBucket = "pivotalnetwork"
	defaultRegion = "eu-west-1"
)

type OutCommand struct {
	binaryVersion   string
	logger          logger.Logger
	outDir          string
	sourcesDir      string
	logFilePath     string
	s3OutBinaryName string
}

type OutCommandConfig struct {
	BinaryVersion   string
	Logger          logger.Logger
	OutDir          string
	SourcesDir      string
	LogFilePath     string
	S3OutBinaryName string
}

func NewOutCommand(config OutCommandConfig) *OutCommand {
	return &OutCommand{
		binaryVersion:   config.BinaryVersion,
		logger:          config.Logger,
		outDir:          config.OutDir,
		sourcesDir:      config.SourcesDir,
		logFilePath:     config.LogFilePath,
		s3OutBinaryName: config.S3OutBinaryName,
	}
}

func (c *OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	if c.outDir == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "out dir")
	}

	if input.Source.APIToken == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "api_token")
	}

	if input.Source.ProductSlug == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "product_slug")
	}

	if input.Params.VersionFile == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "version_file")
	}

	if input.Params.ReleaseTypeFile == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "release_type_file")
	}

	if input.Params.EulaSlugFile == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "eula_slug_file")
	}

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	if !skipUpload {
		if input.Source.AccessKeyID == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "access_key_id")
		}

		if input.Source.SecretAccessKey == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "secret_access_key")
		}

		if input.Params.FileGlob == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "file glob")
		}

		if input.Params.FilepathPrefix == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "s3_filepath_prefix")
		}
	}

	c.logger.Debugf("Received input: %+v\n", input)

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	productSlug := input.Source.ProductSlug

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(c.binaryVersion, "put", productSlug),
	}
	pivnetClient := pivnet.NewClient(
		clientConfig,
		c.logger,
	)

	config := pivnet.CreateReleaseConfig{
		ProductSlug:     productSlug,
		ReleaseType:     readStringContents(c.sourcesDir, input.Params.ReleaseTypeFile),
		EulaSlug:        readStringContents(c.sourcesDir, input.Params.EulaSlugFile),
		ProductVersion:  readStringContents(c.sourcesDir, input.Params.VersionFile),
		Description:     readStringContents(c.sourcesDir, input.Params.DescriptionFile),
		ReleaseNotesURL: readStringContents(c.sourcesDir, input.Params.ReleaseNotesURLFile),
		ReleaseDate:     readStringContents(c.sourcesDir, input.Params.ReleaseDateFile),
	}

	release, err := pivnetClient.CreateRelease(config)
	if err != nil {
		log.Fatalln(err)
	}

	if skipUpload {
		c.logger.Debugf("File glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		bucket := input.Source.Bucket
		if bucket == "" {
			bucket = defaultBucket
		}

		region := input.Source.Region
		if region == "" {
			region = defaultRegion
		}

		logFile, err := os.OpenFile(c.logFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}

		s3Client := s3.NewClient(s3.NewClientConfig{
			AccessKeyID:     input.Source.AccessKeyID,
			SecretAccessKey: input.Source.SecretAccessKey,
			RegionName:      region,
			Bucket:          bucket,

			Logger: c.logger,

			Stdout: os.Stdout,
			Stderr: logFile,

			OutBinaryPath: filepath.Join(c.outDir, c.s3OutBinaryName),
		})

		uploaderClient := uploader.NewClient(uploader.Config{
			FileGlob:       input.Params.FileGlob,
			FilepathPrefix: input.Params.FilepathPrefix,
			SourcesDir:     c.sourcesDir,

			Logger: c.logger,

			Transport: s3Client,
		})

		exactGlobs, err := uploaderClient.ExactGlobs()
		if err != nil {
			log.Fatalln(err)
		}

		for _, exactGlob := range exactGlobs {
			fullFilepath := filepath.Join(c.sourcesDir, exactGlob)
			fileContentsMD5, err := md5.NewFileContentsSummer(fullFilepath).Sum()
			if err != nil {
				log.Fatalln(err)
			}

			remotePath, err := uploaderClient.UploadFile(exactGlob)
			if err != nil {
				return concourse.OutResponse{}, err
			}

			product, err := pivnetClient.FindProductForSlug(productSlug)
			if err != nil {
				log.Fatalln(err)
			}

			filename := filepath.Base(exactGlob)
			c.logger.Debugf(
				"Creating product file: {product_slug: %s, filename: %s, aws_object_key: %s, file_version: %s}\n",
				productSlug,
				filename,
				remotePath,
				release.Version,
			)

			productFile, err := pivnetClient.CreateProductFile(pivnet.CreateProductFileConfig{
				ProductSlug:  productSlug,
				Name:         filename,
				AWSObjectKey: remotePath,
				FileVersion:  release.Version,
				MD5:          fileContentsMD5,
			})
			if err != nil {
				log.Fatalln(err)
			}

			c.logger.Debugf(
				"Adding product file: {product_slug: %s, product_id: %d, filename: %s, product_file_id: %d, release_id: %d}\n",
				productSlug,
				product.ID,
				filename,
				productFile.ID,
				release.ID,
			)

			err = pivnetClient.AddProductFile(product.ID, release.ID, productFile.ID)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	availability := readStringContents(c.sourcesDir, input.Params.AvailabilityFile)
	if availability != "Admins Only" {
		releaseUpdate := pivnet.Release{
			ID:           release.ID,
			Availability: availability,
		}
		release, err = pivnetClient.UpdateRelease(productSlug, releaseUpdate)
		if err != nil {
			log.Fatalln(err)
		}

		if availability == "Selected User Groups Only" {
			userGroupIDs := strings.Split(
				readStringContents(c.sourcesDir, input.Params.UserGroupIDsFile),
				",",
			)

			for _, userGroupIDString := range userGroupIDs {
				userGroupID, err := strconv.Atoi(userGroupIDString)
				if err != nil {
					log.Fatalln(err)
				}

				pivnetClient.AddUserGroup(productSlug, release.ID, userGroupID)
			}
		}
	}

	out := concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: release.Version,
		},
		Metadata: []concourse.Metadata{
			{Name: "release_type", Value: release.ReleaseType},
			{Name: "release_date", Value: release.ReleaseDate},
			{Name: "description", Value: release.Description},
			{Name: "release_notes_url", Value: release.ReleaseNotesURL},
			{Name: "eula_slug", Value: release.Eula.Slug},
			{Name: "availability", Value: release.Availability},
		},
	}

	return out, nil
}

func readStringContents(sourcesDir, file string) string {
	if file == "" {
		return ""
	}
	fullPath := filepath.Join(sourcesDir, file)
	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}
