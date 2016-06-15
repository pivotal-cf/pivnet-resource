package check

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
	"github.com/pivotal-golang/lager"
)

type CheckCommand struct {
	logger         lager.Logger
	binaryVersion  string
	filter         filter.Filter
	pivnetClient   gp.Client
	extendedClient gp.ExtendedClient
}

func NewCheckCommand(
	binaryVersion string,
	logger lager.Logger,
	filter filter.Filter,
	pivnetClient gp.Client,
	extendedClient gp.ExtendedClient,
) *CheckCommand {
	return &CheckCommand{
		logger:         logger,
		binaryVersion:  binaryVersion,
		filter:         filter,
		pivnetClient:   pivnetClient,
		extendedClient: extendedClient,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	c.logger.Debug("Received input", lager.Data{"input": input})

	c.logger.Debug("Getting all valid release types")
	releaseTypes, err := c.pivnetClient.ReleaseTypes()
	if err != nil {
		return nil, err
	}

	releaseTypesPrintable := fmt.Sprintf(
		"['%s']",
		strings.Join(releaseTypes, "', '"),
	)

	c.logger.Debug("All release types", lager.Data{"release_types": releaseTypesPrintable})

	releaseType := input.Source.ReleaseType
	if releaseType != "" && !containsString(releaseTypes, releaseType) {
		return nil, fmt.Errorf(
			"provided release_type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	c.logger.Debug("Getting all product versions")
	productSlug := input.Source.ProductSlug

	releases, err := c.pivnetClient.ReleasesForProductSlug(productSlug)
	if err != nil {
		return nil, err
	}

	if releaseType != "" {
		c.logger.Debug("Filtering all releases by release_type", lager.Data{"release_type": releaseType})

		releases, err = c.filter.ReleasesByReleaseType(releases, releaseType)
		if err != nil {
			return nil, err
		}
	}

	productVersion := input.Source.ProductVersion
	if productVersion != "" {
		c.logger.Debug("Filtering all releases by product_version", lager.Data{"product_version": productVersion})

		releases, err = c.filter.ReleasesByVersion(releases, productVersion)
		if err != nil {
			return nil, err
		}
	}

	// ls := lagershim.NewLagerShim(c.logger)
	// extendedClient := extension.NewExtendedClient(c.pivnetClient, ls)
	filteredVersions, err := versions.ProductVersions(c.extendedClient, productSlug, releases)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("Filtered versions", lager.Data{"filtered_versions": filteredVersions})

	if len(filteredVersions) == 0 {
		return concourse.CheckResponse{}, nil
	}

	newVersions, err := versions.Since(filteredVersions, input.Version.ProductVersion)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	c.logger.Debug("New versions", lager.Data{"new_versions": newVersions})

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

	c.logger.Debug("Returning output", lager.Data{"output:": out})

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
