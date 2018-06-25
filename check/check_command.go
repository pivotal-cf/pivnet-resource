package check

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/versions"
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
}

type CheckCommand struct {
	logger        logger.Logger
	binaryVersion string
	filter        filter
	pivnetClient  pivnetClient
	semverSorter  sorter
	logFilePath   string
}

func NewCheckCommand(
	logger logger.Logger,
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
	c.logger.Info("Received input, starting Check CMD run")

	err := c.removeExistingLogFiles()
	if err != nil {
		return nil, err
	}

	releaseType := input.Source.ReleaseType

	err = c.validateReleaseType(releaseType)
	if err != nil {
		return nil, err
	}

	productSlug := input.Source.ProductSlug

	c.logger.Info("Getting all releases")
	releases, err := c.pivnetClient.ReleasesForProductSlug(productSlug)
	if err != nil {
		return nil, err
	}

	if releaseType != "" {
		c.logger.Info(fmt.Sprintf("Filtering all releases by release type: '%s'", releaseType))
		releases, err = c.filter.ReleasesByReleaseType(
			releases,
			pivnet.ReleaseType(releaseType),
		)
		if err != nil {
			return nil, err
		}
	}

	version := input.Source.ProductVersion
	if version != "" {
		c.logger.Info(fmt.Sprintf("Filtering all releases by product version: '%s'", version))
		releases, err = c.filter.ReleasesByVersion(releases, version)
		if err != nil {
			return nil, err
		}
	}

	if input.Source.SortBy == concourse.SortBySemver {
		c.logger.Info("Sorting all releases by semver")
		releases, err = c.semverSorter.SortBySemver(releases)
		if err != nil {
			return nil, err
		}
	}

	vs, err := releaseVersions(releases)
	if err != nil {
		// Untested because versions.CombineVersionAndFingerprint cannot be forced to return an error.
		return concourse.CheckResponse{}, err
	}

	if len(vs) == 0 {
		return concourse.CheckResponse{}, fmt.Errorf("cannot find specified release")
	}

	c.logger.Info("Gathering new versions")

	newVersions, err := versions.Since(vs, input.Version.ProductVersion)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		// Untested because versions.Reverse cannot be forced to return an error.
		return nil, err
	}

	c.logger.Info(fmt.Sprintf("New versions: %v", reversedVersions))

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: vs[0]})
	}

	c.logger.Info("Finishing check and returning ouput")

	return out, nil
}

func (c *CheckCommand) removeExistingLogFiles() error {
	logDir := filepath.Dir(c.logFilePath)
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "*.log*"))
	if err != nil {
		// This is untested because the only error returned by filepath.Glob is a
		// malformed glob, and this glob is hard-coded to be correct.
		return err
	}

	c.logger.Info(fmt.Sprintf("Located logfiles: %v", existingLogFiles))

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFilePath) {
			c.logger.Info(fmt.Sprintf("Removing existing log file: %s", f))
			err := os.Remove(f)
			if err != nil {
				// This is untested because it is too hard to force os.Remove to return
				// an error.
				return err
			}
		}
	}

	return nil
}

func (c *CheckCommand) validateReleaseType(releaseType string) error {
	c.logger.Info(fmt.Sprintf("Validating release type: '%s'", releaseType))
	releaseTypes, err := c.pivnetClient.ReleaseTypes()
	if err != nil {
		return err
	}

	releaseTypesAsStrings := make([]string, len(releaseTypes))
	for i, r := range releaseTypes {
		releaseTypesAsStrings[i] = string(r)
	}

	if releaseType != "" && !containsString(releaseTypesAsStrings, releaseType) {
		releaseTypesPrintable := fmt.Sprintf("['%s']", strings.Join(releaseTypesAsStrings, "', '"))
		return fmt.Errorf(
			"provided release type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	return nil
}

func containsString(strings []string, str string) bool {
	for _, s := range strings {
		if str == s {
			return true
		}
	}
	return false
}

func releaseVersions(releases []pivnet.Release) ([]string, error) {
	releaseVersions := make([]string, len(releases))

	var err error
	for i, r := range releases {
		releaseVersions[i], err = versions.CombineVersionAndFingerprint(r.Version, r.SoftwareFilesUpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return releaseVersions, nil
}
