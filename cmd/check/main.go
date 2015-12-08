package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource"
)

const (
	url = "https://network.pivotal.io/api/v2"
)

type input struct {
	Source struct {
		APIToken     string `json:"api_token"`
		ResourceName string `json:"resource_name"`
	} `json:"source"`
}

type output []Release

type Release struct {
	Version string `json:"version"`
}

func main() {
	var i input

	err := json.NewDecoder(os.Stdin).Decode(&i)
	if err != nil {
		log.Fatalln(err)
	}

	client := pivnet.NewClient(url, i.Source.APIToken)

	versions, err := client.ProductVersions(i.Source.ResourceName)
	if err != nil {
		log.Fatalln(err)
	}

	var out output
	for _, v := range versions {
		out = append(out, Release{Version: v})
	}

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
