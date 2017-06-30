package main

import (
	"encoding/json"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/check"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/semver"
	"github.com/pivotal-cf/pivnet-resource/sorter"
	"github.com/pivotal-cf/pivnet-resource/uaa"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/pivotal-cf/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
	"io/ioutil"
	"log"
	"os"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

type AuthResp struct {
	Token string `json: "token"`
}

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
	ls := logshim.NewLogShim(logger, logger, verbose)

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

	var usingUAAToken = false
	apiToken := input.Source.APIToken

	if input.Source.Username != "" {
		usingUAAToken = true
		tokenFetcher := uaa.NewTokenFetcher(input.Source.Endpoint, input.Source.Username, input.Source.Password)
		apiToken, err = tokenFetcher.GetToken()

		if err != nil {
			log.Fatalf("Exiting with error: %s", err)
		}
	} else {
		logger.Println("The use of static API tokens is deprecated and will be removed. Please see https://github.com/pivotal-cf/pivnet-resource#source-configuration for details.")
	}

	clientConfig := pivnet.ClientConfig{
		Host:              endpoint,
		Token:             apiToken,
		UserAgent:         useragent.UserAgent(version, "check", input.Source.ProductSlug),
		SkipSSLValidation: input.Source.SkipSSLValidation,
		UsingUAAToken:     usingUAAToken,
	}
	client := gp.NewClient(
		clientConfig,
		ls,
	)

	f := filter.NewFilter(ls)

	semverConverter := semver.NewSemverConverter(ls)
	s := sorter.NewSorter(ls, semverConverter)

	response, err := check.NewCheckCommand(
		ls,
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
