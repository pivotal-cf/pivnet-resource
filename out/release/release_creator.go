package release

import (
	"fmt"
	"log"
	"strings"

	"github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
)

type ReleaseCreator struct {
	pivnet          releaseClient
	metadataFetcher fetcher
	semverConverter semverConverter
	logger          *log.Logger
	metadata        metadata.Metadata
	skipFileCheck   bool
	sourcesDir      string
	productSlug     string
	params          concourse.OutParams
	source          concourse.Source
}

//go:generate counterfeiter --fake-name ReleaseClient . releaseClient
type releaseClient interface {
	EULAs() ([]pivnet.EULA, error)
	ReleaseTypes() ([]string, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	CreateRelease(pivnet.CreateReleaseConfig) (pivnet.Release, error)
	ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error)
}

//go:generate counterfeiter --fake-name FakeSemverConverter . semverConverter
type semverConverter interface {
	ToValidSemver(string) (semver.Version, error)
}

func NewReleaseCreator(
	pivnet releaseClient,
	metadataFetcher fetcher,
	semverConverter semverConverter,
	logger *log.Logger,
	metadata metadata.Metadata,
	skipFileCheck bool,
	params concourse.OutParams,
	source concourse.Source,
	sourcesDir,
	productSlug string,
) ReleaseCreator {
	return ReleaseCreator{
		pivnet:          pivnet,
		metadataFetcher: metadataFetcher,
		semverConverter: semverConverter,
		logger:          logger,
		metadata:        metadata,
		skipFileCheck:   skipFileCheck,
		sourcesDir:      sourcesDir,
		params:          params,
		source:          source,
		productSlug:     productSlug,
	}
}

func (rc ReleaseCreator) Create() (pivnet.Release, error) {
	productVersion := rc.metadataFetcher.Fetch(
		"Version",
		rc.sourcesDir,
		rc.params.VersionFile,
	)

	if rc.source.SortBy == concourse.SortBySemver {
		rc.logger.Println("Resource is configured to sort by semver - checking new version parses as semver")
		v, err := rc.semverConverter.ToValidSemver(productVersion)
		if err != nil {
			return pivnet.Release{}, err
		}
		rc.logger.Printf("Successfully parsed semver as: %s\n", v.String())
	}

	rc.logger.Printf("Getting existing releases for product slug: %s\n", rc.productSlug)
	releases, err := rc.pivnet.ReleasesForProductSlug(rc.productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}

	rc.logger.Println("Mapping existing releases to their versions")
	existingVersions, err := rc.pivnet.ProductVersions(rc.productSlug, releases)
	if err != nil {
		return pivnet.Release{}, err
	}

	for _, v := range existingVersions {
		if v == productVersion {
			return pivnet.Release{},
				fmt.Errorf("release already exists with version: %s", productVersion)
		}
	}

	rc.logger.Println("getting all valid eulas")
	eulas, err := rc.pivnet.EULAs()
	if err != nil {
		return pivnet.Release{}, err
	}

	eulaSlugs := make([]string, len(eulas))
	for i, e := range eulas {
		eulaSlugs[i] = e.Slug
	}

	rc.logger.Println("validating eula_slug")
	eulaSlug := rc.metadataFetcher.Fetch(
		"EULASlug",
		rc.sourcesDir,
		rc.params.EULASlugFile,
	)

	var containsSlug bool
	for _, slug := range eulaSlugs {
		if eulaSlug == slug {
			containsSlug = true
			break
		}
	}

	eulaSlugsPrintable := fmt.Sprintf("['%s']", strings.Join(eulaSlugs, "', '"))

	if !containsSlug {
		return pivnet.Release{}, fmt.Errorf(
			"provided eula_slug: '%s' must be one of: %s",
			eulaSlug,
			eulaSlugsPrintable,
		)
	}

	rc.logger.Println("getting all valid release types")
	releaseTypes, err := rc.pivnet.ReleaseTypes()
	if err != nil {
		return pivnet.Release{}, err
	}

	releaseTypesPrintable := fmt.Sprintf(
		"['%s']",
		strings.Join(releaseTypes, "', '"),
	)

	rc.logger.Println("validating release_type")
	releaseType := rc.metadataFetcher.Fetch(
		"ReleaseType",
		rc.sourcesDir,
		rc.params.ReleaseTypeFile,
	)

	var containsReleaseType bool
	for _, t := range releaseTypes {
		if releaseType == t {
			containsReleaseType = true
			break
		}
	}

	if !containsReleaseType {
		return pivnet.Release{}, fmt.Errorf(
			"provided release_type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	eulaSlug = rc.metadataFetcher.Fetch(
		"EULASlug",
		rc.sourcesDir,
		rc.params.EULASlugFile,
	)

	description := rc.metadataFetcher.Fetch(
		"Description",
		rc.sourcesDir,
		rc.params.DescriptionFile,
	)

	releaseNotesURL := rc.metadataFetcher.Fetch(
		"ReleaseNotesURL",
		rc.sourcesDir,
		rc.params.ReleaseNotesURLFile,
	)

	releaseDate := rc.metadataFetcher.Fetch(
		"ReleaseDate",
		rc.sourcesDir,
		rc.params.ReleaseDateFile,
	)

	config := pivnet.CreateReleaseConfig{
		ProductSlug:     rc.productSlug,
		ReleaseType:     releaseType,
		EULASlug:        eulaSlug,
		ProductVersion:  productVersion,
		Description:     description,
		ReleaseNotesURL: releaseNotesURL,
		ReleaseDate:     releaseDate,
	}
	if rc.metadata.Release != nil {
		config.Controlled = rc.metadata.Release.Controlled
		config.ECCN = rc.metadata.Release.ECCN
		config.LicenseException = rc.metadata.Release.LicenseException
		config.EndOfSupportDate = rc.metadata.Release.EndOfSupportDate
		config.EndOfGuidanceDate = rc.metadata.Release.EndOfGuidanceDate
		config.EndOfAvailabilityDate = rc.metadata.Release.EndOfAvailabilityDate
	}

	rc.logger.Printf("config used to create pivnet release: %+v\n", config)

	rc.logger.Printf("Creating new release")
	return rc.pivnet.CreateRelease(config)
}
