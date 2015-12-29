package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

func main() {
	var input concourse.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := ioutil.TempFile("", "pivnet-resource-check.log")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	logger := logger.NewLogger(sanitizer)

	logDir := filepath.Dir(logFile.Name())
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "pivnet-resource-check.log*"))
	if err != nil {
		log.Fatalln(err)
	}

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(logFile.Name()) {
			logger.Debugf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	mustBeNonEmpty(input.Source.APIToken, "api_token")
	mustBeNonEmpty(input.Source.ProductName, "product_name")

	logger.Debugf("received input: %+v\n", input)

	client := pivnet.NewClient(
		pivnet.URL,
		input.Source.APIToken,
		logger,
	)

	allVersions, err := client.ProductVersions(input.Source.ProductName)
	if err != nil {
		log.Fatalln(err)
	}

	newVersions, err := versions.Since(allVersions, input.Version.ProductVersion)
	if err != nil {
		log.Fatalln(err)
	}

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		log.Fatalln(err)
	}

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: allVersions[0]})
	}

	logger.Debugf("returning output: %+v\n", out)

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
