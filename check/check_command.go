package check

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

//go:generate counterfeiter --fake-name FakeFilter . filter
type filter interface {
	ReleasesByReleaseType(releases []pivnet.Release, releaseType pivnet.ReleaseType) ([]pivnet.Release, error)
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
}

//go:generate counterfeiter --fake-name FakeSorter . sorter
type sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
}

//go:generate counterfeiter --fake-name FakePivnetClient . pivnetClient
type pivnetClient interface {
	ReleaseTypes() ([]pivnet.ReleaseType, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error)
	ReleaseETag(productSlug string, releaseID int) (string, error)
}

type CheckCommand struct {
	logger        *log.Logger
	binaryVersion string
	filter        filter
	pivnetClient  pivnetClient
	semverSorter  sorter
	logFilePath   string
}

func NewCheckCommand(
	logger *log.Logger,
	binaryVersion string,
	filter filter,
	pivnetClient pivnetClient,
	semverSorter sorter,
	logFilePath string,
) *CheckCommand {
	return &CheckCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		filter:        filter,
		pivnetClient:  pivnetClient,
		semverSorter:  semverSorter,
		logFilePath:   logFilePath,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	c.logger.Println("Received input, starting Check CMD run")

	logDir := filepath.Dir(c.logFilePath)
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "*.log*"))
	if err != nil {
		// This is untested because the only error returned by filepath.Glob is a
		// malformed glob, and this glob is hard-coded to be correct.
		return nil, err
	}

	c.logger.Printf("Located logfiles: %v\n", existingLogFiles)

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFilePath) {
			c.logger.Printf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				// This is untested because it is too hard to force os.Remove to return
				// an error.
				return nil, err
			}
		}
	}

	c.logger.Println("Getting all valid release types")
	releaseTypes, err := c.pivnetClient.ReleaseTypes()
	if err != nil {
		return nil, err
	}

	releaseTypesAsStrings := make([]string, len(releaseTypes))
	for i, r := range releaseTypes {
		releaseTypesAsStrings[i] = string(r)
	}

	releaseTypesPrintable := fmt.Sprintf("['%s']", strings.Join(releaseTypesAsStrings, "', '"))

	releaseType := input.Source.ReleaseType
	if releaseType != "" && !containsString(releaseTypesAsStrings, releaseType) {
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
		releases, err = c.filter.ReleasesByReleaseType(
			releases,
			pivnet.ReleaseType(releaseType),
		)
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

	versionsWithETags, err := versions.ProductVersions(c.pivnetClient, productSlug, releases)
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
