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

var (
	l logger.Logger
)

func main() {
	var input concourse.CheckRequest

	logFile, err := ioutil.TempFile("", "pivnet-resource-check.log")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(os.Stderr, "Logging to %s\n", logFile.Name())

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		fmt.Fprintf(logFile, "Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = logger.NewLogger(sanitizer)

	logDir := filepath.Dir(logFile.Name())
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "pivnet-resource-check.log*"))
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(logFile.Name()) {
			l.Debugf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				l.Debugf("Exiting with error: %v\n", err)
				log.Fatalln(err)
			}
		}
	}

	mustBeNonEmpty(input.Source.APIToken, "api_token")
	mustBeNonEmpty(input.Source.ProductSlug, "product_slug")

	l.Debugf("Received input: %+v\n", input)

	clientConfig := pivnet.NewClientConfig{
		URL:       pivnet.URL,
		Token:     input.Source.APIToken,
		UserAgent: "pivnet-resource/dev",
	}
	client := pivnet.NewClient(
		clientConfig,
		l,
	)

	l.Debugf("Getting all product versions\n")

	allVersions, err := client.ProductVersions(input.Source.ProductSlug)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	l.Debugf("All known versions: %+v\n", allVersions)

	newVersions, err := versions.Since(allVersions, input.Version.ProductVersion)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	l.Debugf("New versions: %+v\n", newVersions)

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: allVersions[0]})
	}

	l.Debugf("Returning output: %+v\n", out)

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}

func mustBeNonEmpty(input string, key string) {
	if input == "" {
		l.Debugf("Exiting with error: %s must be provided\n", key)
		log.Fatalf("%s must be provided\n", key)
	}
}
