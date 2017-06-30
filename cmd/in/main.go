package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/downloader"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/in"
	"github.com/pivotal-cf/pivnet-resource/in/filesystem"
	"github.com/pivotal-cf/pivnet-resource/md5sum"
	"github.com/pivotal-cf/pivnet-resource/sha256sum"
	"github.com/pivotal-cf/pivnet-resource/uaa"
	"github.com/pivotal-cf/pivnet-resource/ui"
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

	color.NoColor = false

	logWriter := os.Stderr
	uiPrinter := ui.NewUIPrinter(logWriter)

	logger := log.New(logWriter, "", log.LstdFlags)

	logger.Printf("PivNet Resource version: %s", version)

	if len(os.Args) < 2 {
		uiPrinter.PrintErrorlnf(
			"not enough args - usage: %s <sources directory>",
			os.Args[0],
		)
		os.Exit(1)
	}

	downloadDir := os.Args[1]

	var input concourse.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, logWriter))

	verbose := false
	ls := logshim.NewLogShim(logger, logger, verbose)

	logger.Printf("Creating download directory: %s", downloadDir)

	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	err = validator.NewInValidator(input).Validate()
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
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
	}

	clientConfig := pivnet.ClientConfig{
		Host:              endpoint,
		Token:             apiToken,
		UserAgent:         useragent.UserAgent(version, "get", input.Source.ProductSlug),
		SkipSSLValidation: input.Source.SkipSSLValidation,
		UsingUAAToken:     usingUAAToken,
	}

	client := gp.NewClient(
		clientConfig,
		ls,
	)

	d := downloader.NewDownloader(client, downloadDir, ls, logWriter)

	fs := sha256sum.NewFileSummer()
	md5fs := md5sum.NewFileSummer()

	f := filter.NewFilter(ls)

	fileWriter := filesystem.NewFileWriter(downloadDir, ls)

	response, err := in.NewInCommand(
		ls,
		client,
		f,
		d,
		fs,
		md5fs,
		fileWriter,
	).Run(input)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}
}
