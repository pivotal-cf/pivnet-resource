package filter

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
)

type Filter struct {
	l logger.Logger
}

func NewFilter(l logger.Logger) *Filter {
	return &Filter{
		l: l,
	}
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

type ErrNoMatch struct {
	glob string
}

func (e ErrNoMatch) Error() string {
	return fmt.Sprintf("no match for glob: '%s'", e.glob)
}

func (f Filter) ProductFileNamesByGlobs(
	productFiles []pivnet.ProductFile,
	globs []string,
) ([]pivnet.ProductFile, error) {
	f.l.Debug("filter.ProductFilesNamesByGlobs", logger.Data{"globs": globs})

	filtered := []pivnet.ProductFile{}
	for _, pattern := range globs {
		prevFilteredCount := len(filtered)

		for _, p := range productFiles {
			matched, err := filepath.Match(pattern, p.Name)
			if err != nil {
				return nil, err
			}

			if matched {
				filtered = append(filtered, p)
			}
		}

		if len(filtered) == prevFilteredCount {
			return nil, ErrNoMatch{glob: pattern}
		}
	}

	return filtered, nil
}
