package release

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
)

type metadataFetcher struct {
	metadata      metadata.Metadata
	skipFileCheck bool
}

func NewMetadataFetcher(metadata metadata.Metadata, skipFileCheck bool) metadataFetcher {
	return metadataFetcher{
		skipFileCheck: skipFileCheck,
		metadata:      metadata,
	}
}

func (mf metadataFetcher) Fetch(yamlKey, dir, file string) string {
	if mf.skipFileCheck && mf.metadata.Release != nil {
		metadataValue := reflect.ValueOf(mf.metadata.Release).Elem()
		fieldValue := metadataValue.FieldByName(yamlKey)

		if yamlKey == "UserGroupIDs" {
			var ids []string
			for i := 0; i < fieldValue.Len(); i++ {
				ids = append(ids, fieldValue.Index(i).String())
			}

			return strings.Join(ids, ",")
		}

		return fieldValue.String()
	}

	return readStringContents(dir, file)
}

func readStringContents(sourcesDir, file string) string {
	if file == "" {
		return ""
	}
	fullPath := filepath.Join(sourcesDir, file)
	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}
