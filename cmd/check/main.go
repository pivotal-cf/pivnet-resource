package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/check"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp/lagershim"
	"github.com/pivotal-cf-experimental/pivnet-resource/semver"
	"github.com/pivotal-cf-experimental/pivnet-resource/sorter"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
	"github.com/pivotal-golang/lager"
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
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	pivnetClientLogger := lager.NewLogger("pivnet-resource")
	pivnetClientLogger.RegisterSink(lager.NewWriterSink(sanitizer, lager.DEBUG))
	sp := lagershim.NewLagerShim(pivnetClientLogger)

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

	f := filter.NewFilter()

	extendedClient := gp.NewExtendedClient(*client, sp)
	combinedClient := struct {
		*gp.Client
		*gp.ExtendedClient
	}{
		client,
		extendedClient,
	}

	semverConverter := semver.NewSemverConverter(logger)
	s := sorter.NewSorter(logger, semverConverter)

	response, err := check.NewCheckCommand(
		logger,
		version,
		f,
		combinedClient,
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
