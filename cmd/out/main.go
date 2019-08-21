package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/fatih/color"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/go-pivnet/md5sum"
	"github.com/pivotal-cf/go-pivnet/sha256sum"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/globs"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/s3"
	"github.com/pivotal-cf/pivnet-resource/semver"
	"github.com/pivotal-cf/pivnet-resource/ui"
	"github.com/pivotal-cf/pivnet-resource/uploader"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/pivotal-cf/pivnet-resource/validator"
	"github.com/robdimsdale/sanitizer"
	"github.com/pivotal-cf/go-pivnet/logger"
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

	logger := log.New(logWriter, "", log.LstdFlags|log.Lmicroseconds)

	logger.Printf("PivNet Resource version: %s", version)

	if len(os.Args) < 2 {
		uiPrinter.PrintErrorlnf(
			"not enough args - usage: %s <sources directory>",
			os.Args[0],
		)
		os.Exit(1)
	}

	sourcesDir := os.Args[1]

	outDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	var input concourse.OutRequest
	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, logWriter))

	verbose := input.Source.Verbose
	ls := logshim.NewLogShim(logger, logger, verbose)
	ls.Debug("Verbose output enabled")

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	apiToken := input.Source.APIToken
	token := pivnet.NewAccessTokenOrLegacyToken(apiToken, endpoint, input.Source.SkipSSLValidation, "Pivnet Resource")

	if len(apiToken) < 20 {
		uiPrinter.PrintDeprecationln("The use of static Pivnet API tokens is deprecated and will be removed. Please see https://network.pivotal.io/docs/api#how-to-authenticate for details.")
	}

	client := NewPivnetClientWithToken(
		token,
		endpoint,
		input.Source.SkipSSLValidation,
		useragent.UserAgent(version, "put", input.Source.ProductSlug),
		ls,
	)

	federationToken, err := client.GetFederationToken(input.Source.ProductSlug)
	if err != nil {
		uiPrinter.PrintErrorlnf("Unable to generate Federation Token")
		os.Exit(1)
	}

	s3Client := s3.NewClient(s3.NewClientConfig{
		AccessKeyID:       federationToken.AccessKeyID,
		SecretAccessKey:   federationToken.SecretAccessKey,
		SessionToken:      federationToken.SessionToken,
		RegionName:        federationToken.Region,
		Bucket:            federationToken.Bucket,
		Stderr:            os.Stderr,
		Logger:            ls,
		SkipSSLValidation: input.Source.SkipSSLValidation,
		FileSizeGetter:    s3.FileSizeGetter{},
	})

	prefixFetcher := uploader.NewPrefixFetcher(client, input.Source.ProductSlug)
	filePrefix, err := prefixFetcher.GetPrefix()
	if err != nil {
		uiPrinter.PrintErrorlnf("Could not find product prefix")
		os.Exit(1)
	}

	uploaderClient := uploader.NewClient(uploader.Config{
		FilepathPrefix: 	filePrefix,
		SourcesDir:     	sourcesDir,
		Transport:      	s3Client,
	})

	globber := globs.NewGlobber(globs.GlobberConfig{
		FileGlob:   input.Params.FileGlob,
		SourcesDir: sourcesDir,
		Logger:     ls,
	})

	skipUpload := input.Params.FileGlob == ""

	var m metadata.Metadata
	if input.Params.MetadataFile == "" {
		uiPrinter.PrintErrorlnf("params.metadata_file must be provided")
		os.Exit(1)
	}

	metadataFilepath := filepath.Join(sourcesDir, input.Params.MetadataFile)
	metadataBytes, err := ioutil.ReadFile(metadataFilepath)
	if err != nil {
		uiPrinter.PrintErrorlnf("params.metadata_file could not be read: %s", err.Error())
		os.Exit(1)
	}

	err = yaml.Unmarshal(metadataBytes, &m)
	if err != nil {
		uiPrinter.PrintErrorlnf("params.metadata_file could not be parsed: %s", err.Error())
		os.Exit(1)
	}

	deprecations, err := m.Validate()
	if err != nil {
		uiPrinter.PrintErrorlnf("params.metadata_file is invalid: %s", err.Error())
		os.Exit(1)
	}

	for _, deprecation := range deprecations {
		uiPrinter.PrintDeprecationln(deprecation)
	}

	validation := validator.NewOutValidator(input)
	semverConverter := semver.NewSemverConverter(ls)
	sha256Summer := sha256sum.NewFileSummer()
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
		sha256Summer,
		md5summer,
		m,
		sourcesDir,
		input.Source.ProductSlug,
		asyncTimeout,
		pollFrequency,
		input.Params.SkipProductFilePolling,
	)

	releaseUserGroupsUpdater := release.NewUserGroupsUpdater(
		ls,
		client,
		m,
		input.Source.ProductSlug,
	)

	releaseFileGroupsAdder := release.NewReleaseFileGroupsAdder(
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

	upgradePathSpecifiersCreator := release.NewUpgradePathSpecifiersCreator(
		ls,
		client,
		m,
		input.Source.ProductSlug,
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
		Logger:                       ls,
		OutDir:                       outDir,
		SourcesDir:                   sourcesDir,
		GlobClient:                   globber,
		Validation:                   validation,
		Creator:                      releaseCreator,
		Uploader:                     releaseUploader,
		UserGroupsUpdater:            releaseUserGroupsUpdater,
		ReleaseFileGroupsAdder:       releaseFileGroupsAdder,
		ReleaseDependenciesAdder:     releaseDependenciesAdder,
		DependencySpecifiersCreator:  dependencySpecifiersCreator,
		ReleaseUpgradePathsAdder:     releaseUpgradePathsAdder,
		UpgradePathSpecifiersCreator: upgradePathSpecifiersCreator,
		Finalizer:                    releaseFinalizer,
		M:                            m,
		SkipUpload:                   skipUpload,
	})

	response, err := outCmd.Run(input)
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

func NewPivnetClientWithToken(token pivnet.AccessTokenOrLegacyToken, host string, skipSSLValidation bool, userAgent string, logger logger.Logger) *gp.Client {
	clientConfig := pivnet.ClientConfig{
		Host:              host,
		UserAgent:         userAgent,
		SkipSSLValidation: skipSSLValidation,
	}

	return gp.NewClient(
		token,
		clientConfig,
		logger,
	)
}