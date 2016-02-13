package main

import (
	"encoding/json"
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
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	var input concourse.InRequest
	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf(
			"not enough args - usage: %s <sources directory>", os.Args[0]))
	}

	downloadDir := os.Args[1]

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := ioutil.TempFile("", "pivnet-resource-in.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "PivNet Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l := logger.NewLogger(sanitizer)

	token := input.Source.APIToken
	mustBeNonEmpty(token, "api_token")

	l.Debugf("Received input: %+v\n", input)

	l.Debugf("Creating download directory: %s\n", downloadDir)
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s\n", err.Error())
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     input.Source.APIToken,
		UserAgent: fmt.Sprintf("pivnet-resource/%s", version),
	}
	client := pivnet.NewClient(
		clientConfig,
		l,
	)

	productVersion := input.Version.ProductVersion
	productSlug := input.Source.ProductSlug

	l.Debugf(
		"Getting release: {product_slug: %s, product_version: %s}\n",
		productSlug,
		productVersion,
	)

	release, err := client.GetRelease(productSlug, productVersion)
	if err != nil {
		log.Fatalf("Failed to get Release: %s\n", err.Error())
	}

	l.Debugf(
		"Accepting EULA: {product_slug: %s, release_id: %d}\n",
		productSlug,
		release.ID,
	)

	err = client.AcceptEULA(productSlug, release.ID)
	if err != nil {
		log.Fatalf("EULA acceptance failed for the release: %s\n", err.Error())
	}

	l.Debugf(
		"Getting product files: {release_id: %d}\n",
		release.ID,
	)

	productFiles, err := client.GetProductFiles(release)
	if err != nil {
		log.Fatalf("Failed to get Product Files: %s\n", err.Error())
	}

	l.Debugf(
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
		l.Debugf(
			"Filtering download links with globs: {globs: %+v}\n",
			input.Params.Globs,
		)

		var err error
		downloadLinks, err = filter.DownloadLinksByGlob(downloadLinks, input.Params.Globs)
		if err != nil {
			log.Fatalf("Failed to filter Product Files: %s\n", err.Error())
		}

		l.Debugf(
			"Downloading files: {download_links: %+v, download_dir: %s}\n",
			downloadLinks,
			downloadDir,
		)

		files, err := downloader.Download(downloadDir, downloadLinks, token)
		if err != nil {
			log.Fatalf("Failed to Download Files: %s\n", err.Error())
		}

		for _, f := range files {
			downloadPath := filepath.Join(downloadDir, f)

			l.Debugf(
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

			l.Debugf(
				"MD5 for downloaded file: %s matched expected: %s\n",
				downloadPath,
				md5,
			)
		}
	}

	versionFilepath := filepath.Join(downloadDir, "version")

	l.Debugf(
		"Writing version to file: {version: %s, version_filepath: %s}\n",
		version,
		versionFilepath,
	)

	err = ioutil.WriteFile(versionFilepath, []byte(productVersion), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	out := concourse.InResponse{
		Version: concourse.Version{
			ProductVersion: productVersion,
		},
		Metadata: []concourse.Metadata{
			{Name: "release_type", Value: release.ReleaseType},
			{Name: "release_date", Value: release.ReleaseDate},
			{Name: "description", Value: release.Description},
			{Name: "release_notes_url", Value: release.ReleaseNotesURL},
			{Name: "eula_slug", Value: release.Eula.Slug},
		},
	}

	l.Debugf("Returning output: %+v\n", out)

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}

func mustBeNonEmpty(input string, key string) {
	if input == "" {
		log.Fatalf("%s must be provided\n", key)
	}
}
