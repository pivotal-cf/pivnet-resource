package filter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

func DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error) {
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

func DownloadLinks(p pivnet.ProductFiles) map[string]string {
	links := make(map[string]string)

	for _, productFile := range p.ProductFiles {
		parts := strings.Split(productFile.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		links[fileName] = productFile.Links.Download["href"]
	}

	return links
}

func ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error) {
	var filteredReleases []pivnet.Release

	for _, release := range releases {
		if release.ReleaseType == releaseType {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}
