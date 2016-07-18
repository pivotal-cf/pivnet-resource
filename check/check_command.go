package check

import (
	"fmt"
	"log"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

//go:generate counterfeiter . Filter
type Filter interface {
	ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error)
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
}

//go:generate counterfeiter . Sorter
type Sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
}

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	ReleaseTypes() ([]string, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
}

//go:generate counterfeiter . ExtendedPivnetClient
type ExtendedPivnetClient interface {
	ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error)
	ReleaseETag(productSlug string, releaseID int) (string, error)
}

type CheckCommand struct {
	logger         *log.Logger
	binaryVersion  string
	filter         Filter
	pivnetClient   PivnetClient
	extendedClient ExtendedPivnetClient
	semverSorter   Sorter
}

func NewCheckCommand(
	logger *log.Logger,
	binaryVersion string,
	filter Filter,
	pivnetClient PivnetClient,
	extendedClient ExtendedPivnetClient,
	semverSorter Sorter,
) *CheckCommand {
	return &CheckCommand{
		logger:         logger,
		binaryVersion:  binaryVersion,
		filter:         filter,
		pivnetClient:   pivnetClient,
		extendedClient: extendedClient,
		semverSorter:   semverSorter,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	c.logger.Println("Received input, starting Check CMD run")

	c.logger.Println("Getting all valid release types")
	releaseTypes, err := c.pivnetClient.ReleaseTypes()
	if err != nil {
		return nil, err
	}

	releaseTypesPrintable := fmt.Sprintf("['%s']", strings.Join(releaseTypes, "', '"))

	releaseType := input.Source.ReleaseType
	if releaseType != "" && !containsString(releaseTypes, releaseType) {
		return nil, fmt.Errorf(
			"provided release_type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	productSlug := input.Source.ProductSlug

	c.logger.Println("Getting all product versions")
	releases, err := c.pivnetClient.ReleasesForProductSlug(productSlug)
	if err != nil {
		return nil, err
	}

	if releaseType != "" {
		c.logger.Println("Filtering all releases by release_type")
		releases, err = c.filter.ReleasesByReleaseType(releases, releaseType)
		if err != nil {
			return nil, err
		}
	}

	productVersion := input.Source.ProductVersion
	if productVersion != "" {
		c.logger.Println("Filtering all releases by product_version")
		releases, err = c.filter.ReleasesByVersion(releases, productVersion)
		if err != nil {
			return nil, err
		}
	}

	if input.Source.SortBy == concourse.SortBySemver {
		c.logger.Println("Sorting all releases by semver")
		releases, err = c.semverSorter.SortBySemver(releases)
		if err != nil {
			return nil, err
		}
	}

	versionsWithETags, err := versions.ProductVersions(c.extendedClient, productSlug, releases)
	if err != nil {
		return nil, err
	}

	if len(versionsWithETags) == 0 {
		return concourse.CheckResponse{}, nil
	}

	c.logger.Println("Gathering new versions")

	newVersions, err := versions.Since(versionsWithETags, input.Version.ProductVersion)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	c.logger.Printf("New versions contained: %s", newVersions)

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		// Untested because versions.Reverse cannot be forced to return an error.
		return nil, err
	}

	c.logger.Printf("Reversed versions contained: %s", reversedVersions)

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: versionsWithETags[0]})
	}

	c.logger.Println("Finishing check and returning ouput")

	return out, nil
}

func containsString(strings []string, str string) bool {
	for _, s := range strings {
		if str == s {
			return true
		}
	}
	return false
}
