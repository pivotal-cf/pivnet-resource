package filter

import (
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

func DownloadLinks(p pivnet.ProductFiles) map[string]string {
	links := make(map[string]string)

	for _, productFile := range p.ProductFiles {
		parts := strings.Split(productFile.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		links[fileName] = productFile.Links.Download["href"]
	}

	return links
}
