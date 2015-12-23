package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

const (
	url = "https://network.pivotal.io/api/v2"
)

func main() {
	var input concourse.CheckRequest
	if len(os.Args) < 2 {
		panic("Not enough args")
	}

	downloadDir := os.Args[1]

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	if input.Source.APIToken == "" {
		log.Fatalln("api_token must be provided")
	}

	token := input.Source.APIToken

	client := pivnet.NewClient(url, token)
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	version := input.Version.ProductVersion

	release, err := client.GetRelease(input.Source.ProductName, version)
	if err != nil {
		log.Fatalf("Failed to get Release: %s", err)
	}

	productFiles, err := client.GetProductFiles(release)
	if err != nil {
		log.Fatalf("Failed to get Product Files: %s", err)
	}

	downloadLinks := filter.DownloadLinks(productFiles)

	err = downloader.Download(downloadDir, downloadLinks, token)
	if err != nil {
		log.Fatalf("Failed to Download Files: %s", err)
	}

	out := concourse.InResponse{
		Version: concourse.Version{
			ProductVersion: version,
		},
		Metadata: []concourse.Metadata{},
	}

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
