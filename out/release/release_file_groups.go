package release

import (
	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseFileGroupsAdder struct {
	logger      logger.Logger
	pivnet      releaseFileGroupsAdderClient
	metadata    metadata.Metadata
	productSlug string
}

func NewReleaseFileGroupsAdder(
	logger logger.Logger,
	pivnetClient releaseFileGroupsAdderClient,
	metadata metadata.Metadata,
	productSlug string,
) ReleaseFileGroupsAdder {
	return ReleaseFileGroupsAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name ReleaseFileGroupsAdderClient . releaseFileGroupsAdderClient
type releaseFileGroupsAdderClient interface {
	AddFileGroup(productSlug string, releaseID int, fileGroupID int) error
	CreateFileGroup(config pivnet.CreateFileGroupConfig) (pivnet.FileGroup, error)
}

func (rf ReleaseFileGroupsAdder) AddReleaseFileGroups(release pivnet.Release) error {
	for _, fileGroup := range rf.metadata.FileGroups {
		fileGroupID := fileGroup.ID
		if fileGroupID == 0 {
			g, err := rf.pivnet.CreateFileGroup(pivnet.CreateFileGroupConfig{
				ProductSlug: rf.productSlug,
				Name: fileGroup.Name,
			})

			if err != nil {
				return err
			}

			fileGroupID = g.ID
		}

		err := rf.pivnet.AddFileGroup(rf.productSlug, release.ID, fileGroup.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
