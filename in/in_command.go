package in

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
	"github.com/pivotal-golang/lager"
)

type InCommand struct {
	logger       lager.Logger
	downloadDir  string
	pivnetClient gp.Client
	filter       filter.Filter
	downloader   downloader.Downloader
	fileSummer   md5sum.FileSummer
}

func NewInCommand(
	logger lager.Logger,
	downloadDir string,
	pivnetClient gp.Client,
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
	c.logger.Debug("Received input", lager.Data{"input": input})

	productSlug := input.Source.ProductSlug

	c.logger.Debug("Creating download directory", lager.Data{"download_dir": c.downloadDir})
	err := os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		return concourse.InResponse{}, err
	}

	productVersion, etag, err := versions.SplitIntoVersionAndETag(input.Version.ProductVersion)
	if err != nil {
		c.logger.Debug("Parsing of etag failed; continuing without it")
		productVersion = input.Version.ProductVersion
	}

	c.logger.Debug(
		"Getting release",
		lager.Data{
			"product_slug":    productSlug,
			"product_version": productVersion,
			"etag":            etag,
		},
	)

	release, err := c.pivnetClient.GetRelease(productSlug, productVersion)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debug("Release", lager.Data{"release": release})

	c.logger.Debug(
		"Accepting EULA",
		lager.Data{
			"product_slug": productSlug,
			"release_id":   release.ID,
		},
	)

	err = c.pivnetClient.AcceptEULA(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debug("Getting product files", lager.Data{"release_id": release.ID})

	productFiles, err := c.pivnetClient.GetProductFiles(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debug("Found product files", lager.Data{"product_files": productFiles})

	c.logger.Debug("Getting release dependencies", lager.Data{"release_id": release.ID})

	releaseDependencies, err := c.pivnetClient.ReleaseDependencies(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debug("Found release dependencies", lager.Data{"release_dependencies": releaseDependencies})

	err = c.downloadFiles(
		input.Params.Globs,
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

	for _, pf := range productFiles {
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

	for _, d := range releaseDependencies {
		mdata.Dependencies = append(mdata.Dependencies, metadata.Dependency{
			Release: metadata.DependentRelease{
				ID:      d.Release.ID,
				Version: d.Release.Version,
				Product: metadata.Product{
					ID:   d.Release.Product.ID,
					Name: d.Release.Product.Name,
				},
			},
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
	productFiles []pivnet.ProductFile,
	productSlug string,
	releaseID int,
) error {
	c.logger.Debug(
		"Getting download links",
		lager.Data{
			"product_files": productFiles,
		},
	)

	downloadLinks := c.filter.DownloadLinks(productFiles)

	if len(globs) > 0 {
		c.logger.Debug(
			"Filtering download links with globs",
			lager.Data{
				"globs": globs,
			},
		)

		var err error
		downloadLinks, err = c.filter.DownloadLinksByGlob(downloadLinks, globs)
		if err != nil {
			return err
		}

		c.logger.Debug(
			"Downloading files",
			lager.Data{
				"download_links": downloadLinks,
			},
		)

		files, err := c.downloader.Download(downloadLinks)
		if err != nil {
			return err
		}

		fileMD5s := map[string]string{}
		for _, p := range productFiles {
			productFile, err := c.pivnetClient.GetProductFile(
				productSlug,
				releaseID,
				p.ID,
			)
			if err != nil {
				return err
			}

			parts := strings.Split(productFile.AWSObjectKey, "/")

			if len(parts) < 1 {
				panic("not enough components to form filename")
			}

			fileName := parts[len(parts)-1]

			if fileName == "" {
				panic("empty file name")
			}

			fileMD5s[fileName] = productFile.MD5
		}

		c.logger.Debug("All file MD5s", lager.Data{"md5s": fileMD5s})

		err = c.compareMD5s(files, fileMD5s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c InCommand) writeVersionFile(versionWithETag string) error {
	versionFilepath := filepath.Join(c.downloadDir, "version")

	c.logger.Debug(
		"Writing version to file",
		lager.Data{
			"version_with_etag": versionWithETag,
			"version_filepath":  versionFilepath,
		},
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
	c.logger.Debug(
		"Writing metadata to json file",
		lager.Data{
			"metadata": mdata,
			"filepath": jsonMetadataFilepath,
		},
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
	c.logger.Debug(
		"Writing metadata to json file",
		lager.Data{
			"metadata": mdata,
			"filepath": yamlMetadataFilepath,
		},
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

func (c InCommand) compareMD5s(filepaths []string, expectedMD5s map[string]string) error {
	for _, downloadPath := range filepaths {
		_, f := filepath.Split(downloadPath)
		c.logger.Debug(
			"Calcuating MD5 for downloaded file",
			lager.Data{
				"path": downloadPath,
			},
		)
		md5, err := c.fileSummer.SumFile(downloadPath)
		if err != nil {
			return err
		}

		expectedMD5 := expectedMD5s[f]
		if md5 != expectedMD5 {
			return fmt.Errorf(
				"Failed MD5 comparison for file: %s. Expected %s, got %s\n",
				f,
				expectedMD5,
				md5,
			)
		}

		c.logger.Debug(
			"MD5 for downloaded file matched expected",
			lager.Data{
				"path": downloadPath,
				"md5":  md5,
			},
		)
	}

	return nil
}
