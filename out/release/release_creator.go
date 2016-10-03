package release

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/blang/semver"
	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseCreator struct {
	pivnet          releaseClient
	semverConverter semverConverter
	logger          *log.Logger
	metadata        metadata.Metadata
	sourcesDir      string
	productSlug     string
	params          concourse.OutParams
	source          concourse.Source
}

//go:generate counterfeiter --fake-name ReleaseClient . releaseClient
type releaseClient interface {
	EULAs() ([]pivnet.EULA, error)
	ReleaseTypes() ([]pivnet.ReleaseType, error)
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
	semverConverter semverConverter,
	logger *log.Logger,
	metadata metadata.Metadata,
	params concourse.OutParams,
	source concourse.Source,
	sourcesDir,
	productSlug string,
) ReleaseCreator {
	return ReleaseCreator{
		pivnet:          pivnet,
		semverConverter: semverConverter,
		logger:          logger,
		metadata:        metadata,
		sourcesDir:      sourcesDir,
		params:          params,
		source:          source,
		productSlug:     productSlug,
	}
}

func (rc ReleaseCreator) Create() (pivnet.Release, error) {
	productVersion := rc.metadata.Release.Version

	if rc.source.SortBy == concourse.SortBySemver {
		rc.logger.Println("Resource is configured to sort by semver - checking new version parses as semver")
		v, err := rc.semverConverter.ToValidSemver(productVersion)
		if err != nil {
			return pivnet.Release{}, err
		}
		rc.logger.Printf("Successfully parsed semver as: %s\n", v.String())
	}

	if rc.source.ProductVersion != "" {
		rc.logger.Printf(
			"validating product_version: '%s' against regex: '%s'",
			productVersion,
			rc.source.ProductVersion,
		)

		match, err := regexp.MatchString(rc.source.ProductVersion, productVersion)
		if err != nil {
			return pivnet.Release{}, err
		}

		if !match {
			return pivnet.Release{}, fmt.Errorf(
				"provided product_version: '%s' does not match regex in source: '%s'",
				productVersion,
				rc.source.ProductVersion,
			)
		}
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
	eulaSlug := rc.metadata.Release.EULASlug

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

	releaseTypesAsStrings := make([]string, len(releaseTypes))
	for i, r := range releaseTypes {
		releaseTypesAsStrings[i] = string(r)
	}

	releaseTypesPrintable := fmt.Sprintf(
		"['%s']",
		strings.Join(releaseTypesAsStrings, "', '"),
	)

	rc.logger.Println("validating release_type")
	releaseType := pivnet.ReleaseType(rc.metadata.Release.ReleaseType)

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

	if pivnet.ReleaseType(rc.source.ReleaseType) != "" &&
		pivnet.ReleaseType(rc.source.ReleaseType) != releaseType {
		return pivnet.Release{}, fmt.Errorf(
			"provided release_type: '%s' must match '%s' from source configuration",
			releaseType,
			rc.source.ReleaseType,
		)
	}

	config := pivnet.CreateReleaseConfig{
		ProductSlug:           rc.productSlug,
		ReleaseType:           string(releaseType),
		EULASlug:              eulaSlug,
		ProductVersion:        productVersion,
		Description:           rc.metadata.Release.Description,
		ReleaseNotesURL:       rc.metadata.Release.ReleaseNotesURL,
		ReleaseDate:           rc.metadata.Release.ReleaseDate,
		Controlled:            rc.metadata.Release.Controlled,
		ECCN:                  rc.metadata.Release.ECCN,
		LicenseException:      rc.metadata.Release.LicenseException,
		EndOfSupportDate:      rc.metadata.Release.EndOfSupportDate,
		EndOfGuidanceDate:     rc.metadata.Release.EndOfGuidanceDate,
		EndOfAvailabilityDate: rc.metadata.Release.EndOfAvailabilityDate,
	}

	rc.logger.Printf("config used to create pivnet release: %+v\n", config)

	rc.logger.Printf("Creating new release")
	return rc.pivnet.CreateRelease(config)
}
