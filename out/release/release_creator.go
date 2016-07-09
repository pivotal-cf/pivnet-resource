package release

import (
	"fmt"
	"log"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

type ReleaseCreator struct {
	pivnet          releaseClient
	metadataFetcher fetcher
	logger          *log.Logger
	metadata        metadata.Metadata
	skipFileCheck   bool
	sourcesDir      string
	productSlug     string
	params          concourse.OutParams
}

//go:generate counterfeiter --fake-name ReleaseClient . releaseClient
type releaseClient interface {
	EULAs() ([]pivnet.EULA, error)
	ReleaseTypes() ([]string, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	CreateRelease(pivnet.CreateReleaseConfig) (pivnet.Release, error)
	ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error)
}

func NewReleaseCreator(pivnet releaseClient, metadataFetcher fetcher, logger *log.Logger, metadata metadata.Metadata, skipFileCheck bool, params concourse.OutParams, sourcesDir, productSlug string) ReleaseCreator {
	return ReleaseCreator{
		pivnet:          pivnet,
		metadataFetcher: metadataFetcher,
		logger:          logger,
		metadata:        metadata,
		skipFileCheck:   skipFileCheck,
		sourcesDir:      sourcesDir,
		params:          params,
		productSlug:     productSlug,
	}
}

func (rc ReleaseCreator) Create() (pivnet.Release, error) {
	productVersion := rc.metadataFetcher.Fetch("Version", rc.sourcesDir, rc.params.VersionFile)

	releases, err := rc.pivnet.ReleasesForProductSlug(rc.productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}

	existingVersions, err := rc.pivnet.ProductVersions(rc.productSlug, releases)
	if err != nil {
		return pivnet.Release{}, err
	}

	for _, v := range existingVersions {
		if v == productVersion {
			return pivnet.Release{}, fmt.Errorf("release already exists with version: %s", productVersion)
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

	eulaSlug := rc.metadataFetcher.Fetch("EULASlug", rc.sourcesDir, rc.params.EULASlugFile)

	var containsSlug bool
	for _, slug := range eulaSlugs {
		if eulaSlug == slug {
			containsSlug = true
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

	releaseTypesPrintable := fmt.Sprintf("['%s']", strings.Join(releaseTypes, "', '"))

	rc.logger.Println("validating release_type")

	releaseType := rc.metadataFetcher.Fetch("ReleaseType", rc.sourcesDir, rc.params.ReleaseTypeFile)

	var containsReleaseType bool
	for _, t := range releaseTypes {
		if releaseType == t {
			containsReleaseType = true
		}
	}

	if !containsReleaseType {
		return pivnet.Release{}, fmt.Errorf(
			"provided release_type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	config := pivnet.CreateReleaseConfig{
		ProductSlug:     rc.productSlug,
		ReleaseType:     releaseType,
		EULASlug:        rc.metadataFetcher.Fetch("EULASlug", rc.sourcesDir, rc.params.EULASlugFile),
		ProductVersion:  productVersion,
		Description:     rc.metadataFetcher.Fetch("Description", rc.sourcesDir, rc.params.DescriptionFile),
		ReleaseNotesURL: rc.metadataFetcher.Fetch("ReleaseNotesURL", rc.sourcesDir, rc.params.ReleaseNotesURLFile),
		ReleaseDate:     rc.metadataFetcher.Fetch("ReleaseDate", rc.sourcesDir, rc.params.ReleaseDateFile),
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

	return rc.pivnet.CreateRelease(config)
}
