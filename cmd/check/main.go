package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/check"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/semver"
	"github.com/pivotal-cf/pivnet-resource/sorter"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/pivotal-cf/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	var input concourse.CheckRequest

	logFile, err := ioutil.TempFile("", "pivnet-check.log")
	if err != nil {
		log.Printf("could not create log file")
	}

	logger := log.New(logFile, "", log.LstdFlags)

	logger.Printf("PivNet Resource version: %s", version)

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, logFile))

	verbose := false
	sp := logshim.NewLogShim(logger, logger, verbose)

	err = validator.NewCheckValidator(input).Validate()
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	clientConfig := pivnet.ClientConfig{
		Host:      endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(version, "check", input.Source.ProductSlug),
	}
	client := gp.NewClient(
		clientConfig,
		sp,
	)

	f := filter.NewFilter(sp)

	semverConverter := semver.NewSemverConverter(logger)
	s := sorter.NewSorter(logger, semverConverter)

	response, err := check.NewCheckCommand(
		logger,
		version,
		f,
		client,
		s,
		logFile.Name(),
	).Run(input)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}
}
