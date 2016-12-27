package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/globs"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/md5sum"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/s3"
	"github.com/pivotal-cf/pivnet-resource/semver"
	"github.com/pivotal-cf/pivnet-resource/uploader"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/pivotal-cf/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

const (
	defaultBucket = "pivotalnetwork"
	defaultRegion = "eu-west-1"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	logger := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

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
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, os.Stderr))

	verbose := false
	ls := logshim.NewLogShim(logger, logger, verbose)

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	clientConfig := pivnet.ClientConfig{
		Host:      endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(version, "put", input.Source.ProductSlug),
	}

	client := gp.NewClient(
		clientConfig,
		ls,
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
		Stderr:          os.Stderr,
		Logger:          ls,
	})

	uploaderClient := uploader.NewClient(uploader.Config{
		FilepathPrefix: input.Params.FilepathPrefix,
		SourcesDir:     sourcesDir,
		Transport:      s3Client,
	})

	globber := globs.NewGlobber(globs.GlobberConfig{
		FileGlob:   input.Params.FileGlob,
		SourcesDir: sourcesDir,
		Logger:     ls,
	})

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	var m metadata.Metadata
	if input.Params.MetadataFile == "" {
		log.Fatalf("params.metadata_file must be provided")
	}

	metadataFilepath := filepath.Join(sourcesDir, input.Params.MetadataFile)
	metadataBytes, err := ioutil.ReadFile(metadataFilepath)
	if err != nil {
		log.Fatalf("params.metadata_file could not be read: %s", err.Error())
	}

	err = yaml.Unmarshal(metadataBytes, &m)
	if err != nil {
		log.Fatalf("params.metadata_file could not be parsed: %s", err.Error())
	}

	err = m.Validate()
	if err != nil {
		log.Fatalf("params.metadata_file is invalid: %s", err.Error())
	}

	validation := validator.NewOutValidator(input)
	semverConverter := semver.NewSemverConverter(ls)
	md5summer := md5sum.NewFileSummer()

	f := filter.NewFilter(ls)

	releaseCreator := release.NewReleaseCreator(
		client,
		semverConverter,
		ls,
		m,
		input.Params,
		input.Source,
		sourcesDir,
		input.Source.ProductSlug,
	)

	asyncTimeout := 1 * time.Hour
	pollFrequency := 5 * time.Second
	releaseUploader := release.NewReleaseUploader(
		uploaderClient,
		client,
		ls,
		md5summer,
		m,
		sourcesDir,
		input.Source.ProductSlug,
		asyncTimeout,
		pollFrequency,
	)

	releaseUserGroupsUpdater := release.NewUserGroupsUpdater(
		ls,
		client,
		m,
		input.Source.ProductSlug,
	)

	releaseDependenciesAdder := release.NewReleaseDependenciesAdder(
		ls,
		client,
		m,
		input.Source.ProductSlug,
	)

	dependencySpecifiersCreator := release.NewDependencySpecifiersCreator(
		ls,
		client,
		m,
		input.Source.ProductSlug,
	)

	releaseUpgradePathsAdder := release.NewReleaseUpgradePathsAdder(
		ls,
		client,
		m,
		input.Source.ProductSlug,
		f,
	)

	releaseFinalizer := release.NewFinalizer(
		client,
		ls,
		input.Params,
		m,
		sourcesDir,
		input.Source.ProductSlug,
	)

	outCmd := out.NewOutCommand(out.OutCommandConfig{
		Logger:                      ls,
		OutDir:                      outDir,
		SourcesDir:                  sourcesDir,
		GlobClient:                  globber,
		Validation:                  validation,
		Creator:                     releaseCreator,
		Uploader:                    releaseUploader,
		UserGroupsUpdater:           releaseUserGroupsUpdater,
		ReleaseDependenciesAdder:    releaseDependenciesAdder,
		DependencySpecifiersCreator: dependencySpecifiersCreator,
		ReleaseUpgradePathsAdder:    releaseUpgradePathsAdder,
		Finalizer:                   releaseFinalizer,
		M:                           m,
		SkipUpload:                  skipUpload,
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
