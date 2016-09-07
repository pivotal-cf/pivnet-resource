package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/downloader"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/gp/lagershim"
	"github.com/pivotal-cf/pivnet-resource/in"
	"github.com/pivotal-cf/pivnet-resource/in/filesystem"
	"github.com/pivotal-cf/pivnet-resource/md5sum"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/pivotal-cf/pivnet-resource/validator"
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

	logger := log.New(os.Stderr, "", log.LstdFlags)

	logger.Printf("PivNet Resource version: %s", version)

	if len(os.Args) < 2 {
		log.Fatalf("not enough args - usage: %s <sources directory>", os.Args[0])
	}

	downloadDir := os.Args[1]

	var input concourse.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, ioutil.Discard)

	l := lager.NewLogger("pivnet-resource")
	l.RegisterSink(lager.NewWriterSink(sanitizer, lager.DEBUG))
	ls := lagershim.NewLagerShim(l)

	logger.Printf("Creating download directory: %s", downloadDir)

	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	err = validator.NewInValidator(input).Validate()
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
		UserAgent: useragent.UserAgent(version, "get", input.Source.ProductSlug),
	}

	client := gp.NewClient(
		clientConfig,
		ls,
	)

	extendedClient := gp.NewExtendedClient(*client, ls)

	d := downloader.NewDownloader(extendedClient, downloadDir, logger)
	fs := md5sum.NewFileSummer()

	f := filter.NewFilter()

	fileWriter := filesystem.NewFileWriter(downloadDir, logger)

	response, err := in.NewInCommand(
		logger,
		client,
		f,
		d,
		fs,
		fileWriter,
	).Run(input)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}
}
