package in

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/in/filesystem"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

//go:generate counterfeiter . Filter
type Filter interface {
	DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error)
	DownloadLinks(p []pivnet.ProductFile) map[string]string
}

type InCommand struct {
	logger       *log.Logger
	downloadDir  string
	pivnetClient gp.Client
	filter       Filter
	downloader   downloader.Downloader
	fileSummer   md5sum.FileSummer
	fileWriter   filesystem.FileWriter
}

func NewInCommand(
	logger *log.Logger,
	pivnetClient gp.Client,
	filter Filter,
	downloader downloader.Downloader,
	fileSummer md5sum.FileSummer,
	fileWriter filesystem.FileWriter,
) *InCommand {
	return &InCommand{
		logger:       logger,
		pivnetClient: pivnetClient,
		filter:       filter,
		downloader:   downloader,
		fileSummer:   fileSummer,
		fileWriter:   fileWriter,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	productSlug := input.Source.ProductSlug

	productVersion, etag, err := versions.SplitIntoVersionAndETag(input.Version.ProductVersion)
	if err != nil {
		c.logger.Println("Parsing of etag failed; continuing without it")
		productVersion = input.Version.ProductVersion
	}

	c.logger.Printf("Getting release for product_slug %s and product_version %s", productSlug, productVersion)

	release, err := c.pivnetClient.GetRelease(productSlug, productVersion)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Printf("Accepting EULA for release_id %d", release.ID)

	err = c.pivnetClient.AcceptEULA(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Println("Getting product files")

	productFiles, err := c.pivnetClient.GetProductFiles(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	// Get individual product files to obtain metadata that isn't found
	// in the endpoint for all product files.
	for i, p := range productFiles {
		productFiles[i], err = c.pivnetClient.GetProductFile(productSlug, release.ID, p.ID)
		if err != nil {
			return concourse.InResponse{}, err
		}
	}

	c.logger.Println("Getting release dependencies")

	releaseDependencies, err := c.pivnetClient.ReleaseDependencies(productSlug, release.ID)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.downloadFiles(input.Params.Globs, productFiles, productSlug, release.ID)
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

	if release.EULA != nil {
		mdata.Release.EULASlug = release.EULA.Slug
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

	err = c.fileWriter.WriteVersionFile(versionWithETag)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.fileWriter.WriteMetadataYAMLFile(mdata)
	if err != nil {
		return concourse.InResponse{}, err
	}

	err = c.fileWriter.WriteMetadataJSONFile(mdata)
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

func (c InCommand) downloadFiles(globs []string, productFiles []pivnet.ProductFile, productSlug string, releaseID int) error {
	c.logger.Println("Getting download links")

	downloadLinks := c.filter.DownloadLinks(productFiles)

	if len(globs) > 0 {
		c.logger.Println("Filtering download links by glob")

		var err error
		downloadLinks, err = c.filter.DownloadLinksByGlob(downloadLinks, globs)
		if err != nil {
			return err
		}

		c.logger.Println("Downloading files")

		files, err := c.downloader.Download(downloadLinks)
		if err != nil {
			return err
		}

		fileMD5s := map[string]string{}
		for _, p := range productFiles {
			parts := strings.Split(p.AWSObjectKey, "/")

			if len(parts) < 1 {
				panic("not enough components to form filename")
			}

			fileName := parts[len(parts)-1]

			if fileName == "" {
				panic("empty file name")
			}

			fileMD5s[fileName] = p.MD5
		}

		err = c.compareMD5s(files, fileMD5s)
		if err != nil {
			return err
		}
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
	c.logger.Println("Calcuating MD5 for downloaded files")

	for _, downloadPath := range filepaths {
		_, f := filepath.Split(downloadPath)

		md5, err := c.fileSummer.SumFile(downloadPath)
		if err != nil {
			return err
		}

		expectedMD5 := expectedMD5s[f]
		if md5 != expectedMD5 {
			c.logger.Printf("Failed MD5 comparison for file: %s. Expected %s, got %s\n", f, expectedMD5, md5)
			return errors.New("failed comparison")
		}

		c.logger.Println("MD5 for downloaded file matched")
	}

	return nil
}
