package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

const (
	url = "https://network.pivotal.io/api/v2"
)

func main() {
	var input concourse.Request

	if len(os.Args) < 2 {
		panic("Not enough args")
	}

	downloadDir := os.Args[1]

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	client := pivnet.NewClient(url, input.Source.APIToken)
	if err != nil {
		panic(err)
	}

	release, err := client.GetRelease(input.Source.ProductName, input.Version["version"])
	if err != nil {
		panic(err)
	}

	productFiles, err := client.GetProductFiles(release)
	if err != nil {
		panic(err)
	}

	downloadLinks, err := filter.DownloadLinks(productFiles)
	if err != nil {
		panic(err)
	}

	err = downloader.Download(downloadDir, downloadLinks)
	if err != nil {
		panic(err)
	}
}
