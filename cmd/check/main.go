package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/check"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
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

	response, err := check.NewCheckCommand(version, l, logFile).Run(input)
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
