package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp/lagershim"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/in/filesystem"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/sanitizer"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string

	l lager.Logger
)

func main() {
	if version == "" {
		version = "dev"
	}

	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf(
			"not enough args - usage: %s <sources directory>", os.Args[0]))
	}

	downloadDir := os.Args[1]

	var input concourse.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := ioutil.TempFile("", "pivnet-resource-in.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "PivNet Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = lager.NewLogger("pivnet-resource")
	l.RegisterSink(lager.NewWriterSink(sanitizer, lager.DEBUG))

	l.Debug("Creating download directory", lager.Data{"download_dir": downloadDir})
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		l.Error("Exiting with error", err)
		log.Fatalln(err)
	}

	err = validator.NewInValidator(input).Validate()
	if err != nil {
		l.Error("Exiting with error", err)
		log.Fatalln(err)
	}

	d := downloader.NewDownloader(input.Source.APIToken, downloadDir, l)
	fs := md5sum.NewFileSummer()

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
		UserAgent: useragent.UserAgent(version, "get", input.Source.ProductSlug),
	}

	client := gp.NewClient(
		clientConfig,
		ls,
	)

	f := filter.NewFilter()

	fileWriter := filesystem.NewFileWriter(downloadDir, l)

	response, err := in.NewInCommand(
		l,
		client,
		f,
		d,
		fs,
		fileWriter,
	).Run(input)
	if err != nil {
		l.Error("Exiting with error", err)
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		l.Error("Exiting with error", err)
		log.Fatalln(err)
	}
}
