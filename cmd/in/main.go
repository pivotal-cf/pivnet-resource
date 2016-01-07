package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

const (
	url = "https://network.pivotal.io/api/v2"
)

func main() {
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

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	logger := logger.NewLogger(sanitizer)

	token := input.Source.APIToken
	mustBeNonEmpty(token, "api_token")

	logger.Debugf("received input: %+v\n", input)

	client := pivnet.NewClient(
		url,
		token,
		logger,
	)

	productVersion := input.Version.ProductVersion

	release, err := client.GetRelease(input.Source.ProductSlug, productVersion)
	if err != nil {
		log.Fatalf("Failed to get Release: %s\n", err.Error())
	}

	productFiles, err := client.GetProductFiles(release)
	if err != nil {
		log.Fatalf("Failed to get Product Files: %s\n", err.Error())
	}

	downloadLinks := filter.DownloadLinks(productFiles)

	if len(input.Params.Globs) > 0 {
		var err error
		downloadLinks, err = filter.DownloadLinksByGlob(downloadLinks, input.Params.Globs)
		if err != nil {
			log.Fatalf("Failed to filter Product Files: %s\n", err.Error())
		}
	}

	err = downloader.Download(downloadDir, downloadLinks, token)
	if err != nil {
		log.Fatalf("Failed to Download Files: %s\n", err.Error())
	}

	versionFilepath := filepath.Join(downloadDir, "version")

	err = ioutil.WriteFile(versionFilepath, []byte(productVersion), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	out := concourse.InResponse{
		Version: concourse.Version{
			ProductVersion: productVersion,
		},
		Metadata: []concourse.Metadata{},
	}

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
