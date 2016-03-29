package in

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

type InCommand struct {
	logger        logger.Logger
	downloadDir   string
	binaryVersion string
}

func NewInCommand(
	version string,
	logger logger.Logger,
	downloadDir string,
) *InCommand {
	return &InCommand{
		logger:        logger,
		downloadDir:   downloadDir,
		binaryVersion: version,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	token := input.Source.APIToken
	if token == "" {
		return concourse.InResponse{}, fmt.Errorf("%s must be provided", "api_token")
	}

	if input.Source.ProductSlug == "" {
		return concourse.InResponse{}, fmt.Errorf("%s must be provided", "product_slug")
	}

	c.logger.Debugf("Creating download directory: %s\n", c.downloadDir)
	err := os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s\n", err.Error())
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	productSlug := input.Source.ProductSlug

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     token,
		UserAgent: useragent.UserAgent(c.binaryVersion, "get", productSlug),
	}
	client := pivnet.NewClient(
		clientConfig,
		c.logger,
	)

	var productVersion, etag string
	if input.Source.ProductVersion != "" {
		c.logger.Debugf("User configured version %s is being used\n", input.Source.ProductVersion)
		productVersion = input.Source.ProductVersion
	} else {
		productVersion, etag, err = versions.SplitIntoVersionAndETag(input.Version.ProductVersion)
		if err != nil {
			c.logger.Debugf("Parsing of etag failed continuing without it\n")
			productVersion = input.Version.ProductVersion
		}
	}

	c.logger.Debugf(
		"Getting release: {product_slug: %s, product_version: %s, etag: %s}\n",
		productSlug,
		productVersion,
		etag,
	)

	release, err := client.GetRelease(productSlug, productVersion)
	if err != nil {
		log.Fatalf("Failed to get Release: %s\n", err.Error())
	}

	c.logger.Debugf(
		"Accepting EULA: {product_slug: %s, release_id: %d}\n",
		productSlug,
		release.ID,
	)

	err = client.AcceptEULA(productSlug, release.ID)
	if err != nil {
		log.Fatalf("EULA acceptance failed for the release: %s\n", err.Error())
	}

	c.logger.Debugf(
		"Getting product files: {release_id: %d}\n",
		release.ID,
	)

	productFiles, err := client.GetProductFiles(release)
	if err != nil {
		log.Fatalf("Failed to get Product Files: %s\n", err.Error())
	}

	c.logger.Debugf(
		"Getting download links: {product_files: %+v}\n",
		productFiles,
	)

	downloadLinksMD5 := map[string]string{}
	for _, p := range productFiles.ProductFiles {
		productFile, err := client.GetProductFile(
			productSlug,
			release.ID,
			p.ID,
		)
		if err != nil {
			log.Fatalf("Failed to get Product File: %s\n", err.Error())
		}

		parts := strings.Split(productFile.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		downloadLinksMD5[fileName] = productFile.MD5
	}

	downloadLinks := filter.DownloadLinks(productFiles)

	if len(input.Params.Globs) > 0 {
		c.logger.Debugf(
			"Filtering download links with globs: {globs: %+v}\n",
			input.Params.Globs,
		)

		var err error
		downloadLinks, err = filter.DownloadLinksByGlob(downloadLinks, input.Params.Globs)
		if err != nil {
			log.Fatalf("Failed to filter Product Files: %s\n", err.Error())
		}

		c.logger.Debugf(
			"Downloading files: {download_links: %+v, download_dir: %s}\n",
			downloadLinks,
			c.downloadDir,
		)

		files, err := downloader.Download(c.downloadDir, downloadLinks, token)
		if err != nil {
			log.Fatalf("Failed to Download Files: %s\n", err.Error())
		}

		for _, f := range files {
			downloadPath := filepath.Join(c.downloadDir, f)

			c.logger.Debugf(
				"Calcuating MD5 for downloaded file: %s\n",
				downloadPath,
			)
			md5, err := md5.NewFileContentsSummer(downloadPath).Sum()
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

	versionFilepath := filepath.Join(c.downloadDir, "version")

	c.logger.Debugf(
		"Writing version to file: {version: %s, version_filepath: %s}\n",
		productVersion,
		versionFilepath,
	)

	versionWithETag, err := versions.CombineVersionAndETag(productVersion, etag)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(versionFilepath, []byte(versionWithETag), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	metadata := []concourse.Metadata{
		{Name: "release_type", Value: release.ReleaseType},
		{Name: "release_date", Value: release.ReleaseDate},
		{Name: "description", Value: release.Description},
		{Name: "release_notes_url", Value: release.ReleaseNotesURL},
	}

	if release.Eula != nil {
		metadata = append(metadata,
			concourse.Metadata{Name: "eula_slug", Value: release.Eula.Slug},
		)
	}

	out := concourse.InResponse{
		Version: concourse.Version{
			ProductVersion: versionWithETag,
		},
		Metadata: metadata,
	}

	return out, nil
}
