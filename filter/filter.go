package filter

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

func DownloadLinksByGlob(downloadLinks map[string]string, glob []string) (map[string]string, error) {
	filtered := make(map[string]string)

	for _, pattern := range glob {
		for file, downloadLink := range downloadLinks {
			matched, err := filepath.Match(pattern, file)
			if err != nil {
				return nil, err
			}
			if matched {
				filtered[file] = downloadLink
			}
		}
	}

	if len(filtered) == 0 {
		return nil, errors.New("no files match glob")
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
