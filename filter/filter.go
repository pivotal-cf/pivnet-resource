package filter

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pivotal-cf/go-pivnet"
)

type Filter struct {
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f Filter) ReleasesByReleaseType(releases []pivnet.Release, releaseType pivnet.ReleaseType) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		if release.ReleaseType == releaseType {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

// ReleasesByVersion returns all releases that match the provided version regex
func (f Filter) ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		match, err := regexp.MatchString(version, release.Version)
		if err != nil {
			return nil, err
		}

		if match {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

func (f Filter) DownloadLinksByGlobs(
	downloadLinks map[string]string,
	glob []string,
	failOnNoMatch bool,
) (map[string]string, error) {
	filtered := make(map[string]string)

	if glob == nil {
		glob = []string{"*"}
	}

	for _, pattern := range glob {
		prevFilteredCount := len(filtered)

		for file, downloadLink := range downloadLinks {
			matched, err := filepath.Match(pattern, file)
			if err != nil {
				return nil, err
			}
			if matched {
				filtered[file] = downloadLink
			}
		}

		if len(filtered) == prevFilteredCount && failOnNoMatch {
			return nil, fmt.Errorf("no files match glob: %s", pattern)
		}
	}

	return filtered, nil
}

func (f Filter) DownloadLinks(p []pivnet.ProductFile) map[string]string {
	links := make(map[string]string)

	for _, productFile := range p {
		parts := strings.Split(productFile.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		if productFile.Links == nil {
			panic("links not present")
		}

		if productFile.Links.Download == nil {
			panic("download links not present")
		}

		links[fileName] = productFile.Links.Download["href"]
	}

	return links
}
