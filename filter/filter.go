package filter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

//go:generate counterfeiter . Filter

type Filter interface {
	DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error)
	DownloadLinks(p pivnet.ProductFiles) map[string]string
}

type filter struct {
}

func NewFilter() Filter {
	return &filter{}
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

func (f filter) DownloadLinks(p pivnet.ProductFiles) map[string]string {
	links := make(map[string]string)

	for _, productFile := range p.ProductFiles {
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
