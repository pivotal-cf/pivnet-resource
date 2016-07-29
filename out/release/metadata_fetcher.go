package release

import (
	"reflect"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
)

type metadataFetcher struct {
	metadata metadata.Metadata
}

func NewMetadataFetcher(metadata metadata.Metadata) metadataFetcher {
	return metadataFetcher{
		metadata: metadata,
	}
}

func (mf metadataFetcher) Fetch(yamlKey string) string {
	if mf.metadata.Release != nil {
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

	return ""
}
