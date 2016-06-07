package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/globs"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/out"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/s3"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

const (
	s3OutBinaryName = "s3-out"
	defaultBucket   = "pivotalnetwork"
	defaultRegion   = "eu-west-1"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf(
			"not enough args - usage: %s <sources directory>", os.Args[0]))
	}

	sourcesDir := os.Args[1]

	outDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	var input concourse.OutRequest

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := ioutil.TempFile("", "pivnet-resource-out.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "PivNet Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l := logger.NewLogger(sanitizer)

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(version, "put", input.Source.ProductSlug),
	}

	pivnetClient := pivnet.NewClient(
		clientConfig,
		l,
	)

	bucket := input.Source.Bucket
	if bucket == "" {
		bucket = defaultBucket
	}

	region := input.Source.Region
	if region == "" {
		region = defaultRegion
	}

	s3Client := s3.NewClient(s3.NewClientConfig{
		AccessKeyID:     input.Source.AccessKeyID,
		SecretAccessKey: input.Source.SecretAccessKey,
		RegionName:      region,
		Bucket:          bucket,
		Logger:          l,
		Stdout:          os.Stdout,
		Stderr:          logFile,
		OutBinaryPath:   filepath.Join(outDir, s3OutBinaryName),
	})

	uploaderClient := uploader.NewClient(uploader.Config{
		FilepathPrefix: input.Params.FilepathPrefix,
		SourcesDir:     sourcesDir,
		Logger:         l,
		Transport:      s3Client,
	})

	globber := globs.NewGlobber(globs.GlobberConfig{
		FileGlob:   input.Params.FileGlob,
		SourcesDir: sourcesDir,
		Logger:     l,
	})

	validation := validator.NewOutValidator(input)

	outCmd := out.NewOutCommand(out.OutCommandConfig{
		Logger:         l,
		OutDir:         outDir,
		SourcesDir:     sourcesDir,
		ScreenWriter:   log.New(os.Stderr, "", 0),
		PivnetClient:   pivnetClient,
		UploaderClient: uploaderClient,
		GlobClient:     globber,
		Validation:     validation,
	})

	response, err := outCmd.Run(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	l.Debugf("Returning output: %+v\n", response)

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}
