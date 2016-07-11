package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/globs"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/s3"
	"github.com/pivotal-cf-experimental/pivnet-resource/semver"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
	"github.com/pivotal-golang/lager"
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

	logger := log.New(os.Stderr, "pivnet out ", log.LstdFlags)

	logger.Printf("PivNet Resource version: %s", version)

	if len(os.Args) < 2 {
		log.Fatalf("not enough args - usage: %s <sources directory>", os.Args[0])
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

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, ioutil.Discard)

	l := lager.NewLogger("pivnet-resource")
	l.RegisterSink(lager.NewWriterSink(sanitizer, lager.DEBUG))

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
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
		OutBinaryPath:   filepath.Join(outDir, s3OutBinaryName),
	})

	uploaderClient := uploader.NewClient(uploader.Config{
		FilepathPrefix: input.Params.FilepathPrefix,
		SourcesDir:     sourcesDir,
		Transport:      s3Client,
	})

	globber := globs.NewGlobber(globs.GlobberConfig{
		FileGlob:   input.Params.FileGlob,
		SourcesDir: sourcesDir,
		Logger:     logger,
	})

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	var m metadata.Metadata
	var skipFileCheck bool
	if input.Params.MetadataFile != "" {
		metadataFilepath := filepath.Join(sourcesDir, input.Params.MetadataFile)
		metadataBytes, err := ioutil.ReadFile(metadataFilepath)
		if err != nil {
			log.Fatalf("metadata_file could not be read: %s", err)
		}

		err = yaml.Unmarshal(metadataBytes, &m)
		if err != nil {
			log.Fatalf("metadata_file could not be parsed: %s", err)
		}

		err = m.Validate()
		if err != nil {
			log.Fatalf("metadata_file is invalid: %s", err)
		}

		skipFileCheck = true
	}

	validation := validator.NewOutValidator(input)

	metadataFetcher := release.NewMetadataFetcher(m, skipFileCheck)

	semverConverter := semver.NewSemverConverter(logger)

	md5summer := md5sum.NewFileSummer()

	releaseCreator := release.NewReleaseCreator(
		pivnetClient,
		metadataFetcher,
		semverConverter,
		logger,
		m,
		skipFileCheck,
		input.Params,
		input.Source,
		sourcesDir,
		input.Source.ProductSlug,
	)

	releaseUploader := release.NewReleaseUploader(
		uploaderClient,
		pivnetClient,
		logger,
		md5summer,
		m,
		skipUpload,
		sourcesDir,
		input.Source.ProductSlug,
	)

	releaseFinalizer := release.NewFinalizer(
		pivnetClient,
		metadataFetcher,
		input.Params,
		sourcesDir,
		input.Source.ProductSlug,
	)

	outCmd := out.NewOutCommand(out.OutCommandConfig{
		SkipFileCheck: skipFileCheck,
		Logger:        logger,
		OutDir:        outDir,
		SourcesDir:    sourcesDir,
		GlobClient:    globber,
		Validation:    validation,
		Creator:       releaseCreator,
		Uploader:      releaseUploader,
		Finalizer:     releaseFinalizer,
		M:             m,
	})

	response, err := outCmd.Run(input)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		log.Fatalln(err)
	}
}
