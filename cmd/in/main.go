package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

func main() {
	var input concourse.Request

	if len(os.Args) < 2 {
		panic("Not enough args")
	}

	dir := os.Args[1]

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

	// Get all the releases for the product
	releasesURL := c.url + "/products/" + slug + "/releases"
	req, err := http.NewRequest("GET", releasesURL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", input.Source.APIToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	response := pivnet.Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	// Find the right release
	var versions []string
	var id int
	for _, r := range response.Releases {
		if r.Version == input.Version.ProductVersion {
			id = r.ID
		}
	}

	// Find a list of all the files for this release
	releasesURL := c.url + "/products/" + slug + "/releases/" + id + "product_files/download"
	req, err = http.NewRequest("GET", releasesURL, nil)
	if err != nil {
		panic(err)
	}

	// Download each of the files

	// Output something maybe?
	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
