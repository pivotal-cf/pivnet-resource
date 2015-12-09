package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

const (
	url = "https://network.pivotal.io/api/v2"
)

func main() {
	var input concourse.Request

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	client := pivnet.NewClient(url, input.Source.APIToken)

	allVersions, err := client.ProductVersions(input.Source.ResourceName)
	if err != nil {
		log.Fatalln(err)
	}

	newVersions, err := versions.Since(allVersions, input.Version["version"])
	if err != nil {
		log.Fatalln(err)
	}

	var out concourse.Response

	for i := len(newVersions) - 1; i >= 0; i-- {
		v := newVersions[i]
		out = append(out, pivnet.Release{Version: v})
	}

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
