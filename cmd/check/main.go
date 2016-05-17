package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/check"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/lagershim"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string

	l logger.Logger
)

func main() {
	if version == "" {
		version = "dev"
	}

	var input concourse.CheckRequest

	logFile, err := ioutil.TempFile("", "pivnet-resource-check.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "PivNet Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "Logging to %s\n", logFile.Name())

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		fmt.Fprintf(logFile, "Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = logger.NewLogger(sanitizer)

	err = validator.NewCheckValidator(input).Validate()
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	ls := lagershim.NewLagerShim(l)

	clientConfig := pivnet.ClientConfig{
		Host:      endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(version, "check", input.Source.ProductSlug),
	}
	client := gp.NewClient(
		clientConfig,
		ls,
	)

	f := filter.NewFilter()

	extendedClient := gp.NewExtendedClient(client, ls)

	response, err := check.NewCheckCommand(
		version,
		l,
		logFile.Name(),
		f,
		client,
		extendedClient,
	).Run(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}
