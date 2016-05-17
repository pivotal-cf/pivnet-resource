package pivnet_filter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
)

//go:generate counterfeiter . Filter

type Filter interface {
	ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error)
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
	DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error)
	DownloadLinks(p []pivnet.ProductFile) map[string]string
}

type filter struct {
}

func NewFilter() Filter {
	return &filter{}
}

func (f filter) ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		if release.ReleaseType == releaseType {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

func (f filter) ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		if release.Version == version {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

func (f filter) DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error) {
	filtered := make(map[string]string)

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

		if len(filtered) == prevFilteredCount {
			return nil, fmt.Errorf("no files match glob: %s", pattern)
		}
	}

	return filtered, nil
}

func (f filter) DownloadLinks(p []pivnet.ProductFile) map[string]string {
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
