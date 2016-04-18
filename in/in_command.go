package in

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

type InCommand struct {
	logger       logger.Logger
	downloadDir  string
	pivnetClient pivnet.Client
	filter       filter.Filter
	downloader   downloader.Downloader
	fileSummer   md5sum.FileSummer
}

func NewInCommand(
	logger logger.Logger,
	downloadDir string,
	pivnetClient pivnet.Client,
	filter filter.Filter,
	downloader downloader.Downloader,
	fileSummer md5sum.FileSummer,
) *InCommand {
	return &InCommand{
		logger:       logger,
		downloadDir:  downloadDir,
		pivnetClient: pivnetClient,
		filter:       filter,
		downloader:   downloader,
		fileSummer:   fileSummer,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	productSlug := input.Source.ProductSlug

	c.logger.Debugf("Creating download directory: %s\n", c.downloadDir)
	err := os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		return concourse.InResponse{}, err
	}

	productVersion, etag, err := versions.SplitIntoVersionAndETag(input.Version.ProductVersion)
	if err != nil {
		c.logger.Debugf("Parsing of etag failed; continuing without it\n")
		productVersion = input.Version.ProductVersion
	}

	c.logger.Debugf(
		"Getting release: {product_slug: %s, product_version: %s, etag: %s}\n",
		productSlug,
		productVersion,
		etag,
	)

	release, err := c.pivnetClient.GetRelease(productSlug, productVersion)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Release: %+v\n", release)

	c.logger.Debugf(
		"Accepting EULA: {product_slug: %s, release_id: %d}\n",
		productSlug,
		release.ID,
	)

	err = c.pivnetClient.AcceptEULA(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Getting product files: {release_id: %d}\n", release.ID)

	productFiles, err := c.pivnetClient.GetProductFiles(release)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.downloadFiles(
		input.Params.Globs,
		input.Source.APIToken,
		productFiles,
		productSlug,
		release.ID,
	)
	if err != nil {
		return concourse.InResponse{}, err
	}

	versionWithETag, err := versions.CombineVersionAndETag(productVersion, etag)

	mdata := metadata.Metadata{
		Release: &metadata.Release{
			Version:               release.Version,
			ReleaseType:           release.ReleaseType,
			ReleaseDate:           release.ReleaseDate,
			Description:           release.Description,
			ReleaseNotesURL:       release.ReleaseNotesURL,
			Availability:          release.Availability,
			Controlled:            release.Controlled,
			ECCN:                  release.ECCN,
			LicenseException:      release.LicenseException,
			EndOfSupportDate:      release.EndOfSupportDate,
			EndOfGuidanceDate:     release.EndOfGuidanceDate,
			EndOfAvailabilityDate: release.EndOfAvailabilityDate,
		},
	}

	for _, pf := range productFiles.ProductFiles {
		mdata.ProductFiles = append(mdata.ProductFiles, metadata.ProductFile{
			ID:           pf.ID,
			File:         pf.Name,
			Description:  pf.Description,
			AWSObjectKey: pf.AWSObjectKey,
			FileType:     pf.FileType,
			FileVersion:  pf.FileVersion,
			MD5:          pf.MD5,
		})
	}

	err = c.writeVersionFile(versionWithETag)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.writeMetadataYAMLFile(mdata)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.writeMetadataJSONFile(mdata)
	if err != nil {
		return concourse.InResponse{}, err
	}

	concourseMetadata := c.addReleaseMetadata([]concourse.Metadata{}, release)

	out := concourse.InResponse{
		Version: concourse.Version{
			ProductVersion: versionWithETag,
		},
		Metadata: concourseMetadata,
	}

	return out, nil
}

func (c InCommand) downloadFiles(
	globs []string,
	apiToken string,
	productFiles pivnet.ProductFiles,
	productSlug string,
	releaseID int,
) error {
	c.logger.Debugf(
		"Getting download links: {product_files: %+v}\n",
		productFiles,
	)

	downloadLinks := c.filter.DownloadLinks(productFiles)

	if len(globs) > 0 {
		c.logger.Debugf(
			"Filtering download links with globs: {globs: %+v}\n",
			globs,
		)

		var err error
		downloadLinks, err = c.filter.DownloadLinksByGlob(downloadLinks, globs)
		if err != nil {
			return err
		}

		c.logger.Debugf(
			"Downloading files: {download_links: %+v, download_dir: %s}\n",
			downloadLinks,
			c.downloadDir,
		)

		files, err := c.downloader.Download(c.downloadDir, downloadLinks, apiToken)
		if err != nil {
			return err
		}

		downloadLinksMD5 := map[string]string{}
		for _, p := range productFiles.ProductFiles {
			productFile, err := c.pivnetClient.GetProductFile(
				productSlug,
				releaseID,
				p.ID,
			)
			if err != nil {
				return err
			}

			parts := strings.Split(productFile.AWSObjectKey, "/")
			fileName := parts[len(parts)-1]

			downloadLinksMD5[fileName] = productFile.MD5
		}

		for _, f := range files {
			downloadPath := filepath.Join(c.downloadDir, f)

			c.logger.Debugf(
				"Calcuating MD5 for downloaded file: %s\n",
				downloadPath,
			)
			md5, err := c.fileSummer.SumFile(downloadPath)
			if err != nil {
				log.Fatalf("Failed to calculate MD5: %s\n", err.Error())
			}

			expectedMD5 := downloadLinksMD5[f]
			if md5 != expectedMD5 {
				log.Fatalf(
					"Failed MD5 comparison for file: %s. Expected %s, got %s\n",
					f,
					expectedMD5,
					md5,
				)
			}

			c.logger.Debugf(
				"MD5 for downloaded file: %s matched expected: %s\n",
				downloadPath,
				md5,
			)
		}
	}

	return nil
}

func (c InCommand) writeVersionFile(versionWithETag string) error {
	versionFilepath := filepath.Join(c.downloadDir, "version")

	c.logger.Debugf(
		"Writing version to file: {version: %s, version_filepath: %s}\n",
		versionWithETag,
		versionFilepath,
	)

	err := ioutil.WriteFile(versionFilepath, []byte(versionWithETag), os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}

func (c InCommand) writeMetadataJSONFile(mdata metadata.Metadata) error {
	jsonMetadataFilepath := filepath.Join(c.downloadDir, "metadata.json")
	c.logger.Debugf(
		"Writing metadata to json file: {metadata: %+v, metadata_filepath: %s}\n",
		mdata,
		jsonMetadataFilepath,
	)

	jsonMetadata, err := json.Marshal(mdata)
	if err != nil {
		// Untested as it is too hard to force json.Marshal to return an error
		return err
	}

	err = ioutil.WriteFile(jsonMetadataFilepath, jsonMetadata, os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}

func (c InCommand) writeMetadataYAMLFile(mdata metadata.Metadata) error {
	yamlMetadataFilepath := filepath.Join(c.downloadDir, "metadata.yaml")
	c.logger.Debugf(
		"Writing metadata to yaml file: {metadata: %+v, metadata_filepath: %s}\n",
		mdata,
		yamlMetadataFilepath,
	)

	yamlMetadata, err := yaml.Marshal(mdata)
	if err != nil {
		// Untested as it is too hard to force yaml.Marshal to return an error
		return err
	}

	err = ioutil.WriteFile(yamlMetadataFilepath, yamlMetadata, os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}

func (c InCommand) addReleaseMetadata(concourseMetadata []concourse.Metadata, release pivnet.Release) []concourse.Metadata {
	cmdata := append(concourseMetadata,
		concourse.Metadata{Name: "version", Value: release.Version},
		concourse.Metadata{Name: "release_type", Value: release.ReleaseType},
		concourse.Metadata{Name: "release_date", Value: release.ReleaseDate},
		concourse.Metadata{Name: "description", Value: release.Description},
		concourse.Metadata{Name: "release_notes_url", Value: release.ReleaseNotesURL},
		concourse.Metadata{Name: "availability", Value: release.Availability},
		concourse.Metadata{Name: "controlled", Value: fmt.Sprintf("%t", release.Controlled)},
		concourse.Metadata{Name: "eccn", Value: release.ECCN},
		concourse.Metadata{Name: "license_exception", Value: release.LicenseException},
		concourse.Metadata{Name: "end_of_support_date", Value: release.EndOfSupportDate},
		concourse.Metadata{Name: "end_of_guidance_date", Value: release.EndOfGuidanceDate},
		concourse.Metadata{Name: "end_of_availability_date", Value: release.EndOfAvailabilityDate},
	)

	if release.EULA != nil {
		concourseMetadata = append(concourseMetadata,
			concourse.Metadata{Name: "eula_slug", Value: release.EULA.Slug},
		)
	}

	return cmdata
}
