package check

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

type CheckCommand struct {
	logger        logger.Logger
	logFilePath   string
	binaryVersion string
}

func NewCheckCommand(
	binaryVersion string,
	logger logger.Logger,
	logFilePath string,
) *CheckCommand {
	return &CheckCommand{
		logger:        logger,
		logFilePath:   logFilePath,
		binaryVersion: binaryVersion,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	logDir := filepath.Dir(c.logFilePath)
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "pivnet-resource-check.log*"))
	if err != nil {
		// This is untested because the only error returned by filepath.Glob is a
		// malformed glob, and this glob is hard-coded to be correct.
		return nil, err
	}

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFilePath) {
			c.logger.Debugf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				// This is untested because it is too hard to force os.Remove to return
				// an error.
				return nil, err
			}
		}
	}

	if input.Source.APIToken == "" {
		return nil, fmt.Errorf("%s must be provided", "api_token")
	}

	if input.Source.ProductSlug == "" {
		return nil, fmt.Errorf("%s must be provided", "product_slug")
	}

	c.logger.Debugf("Received input: %+v\n", input)

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	productSlug := input.Source.ProductSlug

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     input.Source.APIToken,
		UserAgent: useragent.UserAgent(c.binaryVersion, "check", productSlug),
	}
	client := pivnet.NewClient(
		clientConfig,
		c.logger,
	)

	c.logger.Debugf("Getting all product versions\n")

	releases, err := client.ReleasesForProductSlug(productSlug)
	if err != nil {
		return nil, err
	}

	releaseType := input.Source.ReleaseType
	if releaseType != "" {
		releases, err = filter.ReleasesByReleaseType(releases, releaseType)
		if err != nil {
			return nil, err
		}
	}

	filteredVersions, err := client.ProductVersions(productSlug, releases)
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("Filtered versions: %+v\n", filteredVersions)

	if len(filteredVersions) == 0 {
		return concourse.CheckResponse{}, nil
	}

	newVersions, err := versions.Since(filteredVersions, input.Version.ProductVersion)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	c.logger.Debugf("New versions: %+v\n", newVersions)

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: filteredVersions[0]})
	}

	c.logger.Debugf("Returning output: %+v\n", out)

	return out, nil
}
