package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

func main() {
	var input concourse.Request

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	if input.Source.APIToken == "" {
		log.Fatalln("api_token must be provided")
	}

	if input.Source.ProductName == "" {
		log.Fatalln("product_name must be provided")
	}

	client := pivnet.NewClient(pivnet.URL, input.Source.APIToken)

	allVersions, err := client.ProductVersions(input.Source.ProductName)
	if err != nil {
		log.Fatalln(err)
	}

	newVersions, err := versions.Since(allVersions, input.Version["version"])
	if err != nil {
		log.Fatalln(err)
	}

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		log.Fatalln(err)
	}

	var out concourse.Response
	for _, v := range reversedVersions {
		out = append(out, concourse.Release{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Release{ProductVersion: allVersions[0]})
	}

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
