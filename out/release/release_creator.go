package release

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver"
	pivnet "github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseCreator struct {
	pivnet          releaseClient
	semverConverter semverConverter
	logger          logger.Logger
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
	DeleteRelease(productSlug string, release pivnet.Release) error
}

//go:generate counterfeiter --fake-name FakeSemverConverter . semverConverter
type semverConverter interface {
	ToValidSemver(string) (semver.Version, error)
}

func NewReleaseCreator(
	pivnet releaseClient,
	semverConverter semverConverter,
	logger logger.Logger,
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
	version := rc.metadata.Release.Version

	if rc.source.SortBy == concourse.SortBySemver {
		v, err := rc.semverConverter.ToValidSemver(version)
		if err != nil {
			return pivnet.Release{}, err
		}
		rc.logger.Info(fmt.Sprintf("Successfully parsed semver as: '%s'", v.String()))
	}

	if rc.source.ProductVersion != "" {
		rc.logger.Info(fmt.Sprintf(
			"Validating product version: '%s' against regex: '%s'",
			version,
			rc.source.ProductVersion,
		))

		match, err := regexp.MatchString(rc.source.ProductVersion, version)
		if err != nil {
			return pivnet.Release{}, err
		}

		if !match {
			return pivnet.Release{}, fmt.Errorf(
				"provided product version: '%s' does not match regex in source: '%s'",
				version,
				rc.source.ProductVersion,
			)
		}
	}

	eulaSlug := rc.metadata.Release.EULASlug

	rc.logger.Info(fmt.Sprintf("Validating EULA: '%s'", eulaSlug))

	eulas, err := rc.pivnet.EULAs()
	if err != nil {
		return pivnet.Release{}, err
	}

	eulaSlugs := make([]string, len(eulas))
	for i, e := range eulas {
		eulaSlugs[i] = e.Slug
	}

	var containsSlug bool
	for _, slug := range eulaSlugs {
		if eulaSlug == slug {
			containsSlug = true
			break
		}
	}

	if !containsSlug {
		eulaSlugsPrintable := fmt.Sprintf("['%s']", strings.Join(eulaSlugs, "', '"))
		return pivnet.Release{}, fmt.Errorf(
			"provided EULA slug: '%s' must be one of: %s",
			eulaSlug,
			eulaSlugsPrintable,
		)
	}

	releaseType := pivnet.ReleaseType(rc.metadata.Release.ReleaseType)

	rc.logger.Info(fmt.Sprintf("Validating release type: '%s'", releaseType))

	releaseTypes, err := rc.pivnet.ReleaseTypes()
	if err != nil {
		return pivnet.Release{}, err
	}

	releaseTypesAsStrings := make([]string, len(releaseTypes))
	for i, r := range releaseTypes {
		releaseTypesAsStrings[i] = string(r)
	}

	var containsReleaseType bool
	for _, t := range releaseTypes {
		if releaseType == t {
			containsReleaseType = true
			break
		}
	}

	if !containsReleaseType {
		releaseTypesPrintable := fmt.Sprintf(
			"['%s']",
			strings.Join(releaseTypesAsStrings, "', '"),
		)
		return pivnet.Release{}, fmt.Errorf(
			"provided release type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	if pivnet.ReleaseType(rc.source.ReleaseType) != "" &&
		pivnet.ReleaseType(rc.source.ReleaseType) != releaseType {
		return pivnet.Release{}, fmt.Errorf(
			"provided release type: '%s' must match '%s' from source configuration",
			releaseType,
			rc.source.ReleaseType,
		)
	}

	releases, err := rc.pivnet.ReleasesForProductSlug(rc.productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}

	for _, r := range releases {
		if r.Version == version {
			if rc.params.Override {
				rc.logger.Info(fmt.Sprintf(
					"Deleting existing release: '%s' - id: '%d'",
					r.Version,
					r.ID,
				))

				err := rc.pivnet.DeleteRelease(rc.productSlug, r)
				if err != nil {
					return pivnet.Release{}, err
				}
			} else {
				return pivnet.Release{}, fmt.Errorf(
					"Release '%s' with version '%s' already exists.",
					rc.productSlug,
					version,
				)
			}
		}
	}

	config := pivnet.CreateReleaseConfig{
		ProductSlug:           rc.productSlug,
		ReleaseType:           string(releaseType),
		EULASlug:              eulaSlug,
		Version:               version,
		Description:           rc.metadata.Release.Description,
		ReleaseNotesURL:       rc.metadata.Release.ReleaseNotesURL,
		ReleaseDate:           rc.metadata.Release.ReleaseDate,
		Controlled:            rc.metadata.Release.Controlled,
		ECCN:                  rc.metadata.Release.ECCN,
		LicenseException:      rc.metadata.Release.LicenseException,
		EndOfSupportDate:      rc.metadata.Release.EndOfSupportDate,
		EndOfGuidanceDate:     rc.metadata.Release.EndOfGuidanceDate,
		EndOfAvailabilityDate: rc.metadata.Release.EndOfAvailabilityDate,
		CopyMetadata:          rc.source.CopyMetadata,
	}

	rc.logger.Info(fmt.Sprintf("Creating new release with config: %+v", config))
	release, err := rc.pivnet.CreateRelease(config)
	if err != nil {
		return pivnet.Release{}, err
	}

	rc.logger.Info(fmt.Sprintf("Created new release with ID: %d", release.ID))
	return release, nil
}
