package release

import (
	"fmt"
	pivnet "github.com/pivotal-cf/go-pivnet/v4"
	"github.com/pivotal-cf/go-pivnet/v4/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseImageReferencesAdder struct {
	logger      logger.Logger
	pivnet      releaseImageReferencesAdderClient
	metadata    metadata.Metadata
	productSlug string
}

func NewReleaseImageReferencesAdder(
	logger logger.Logger,
	pivnetClient releaseImageReferencesAdderClient,
	metadata metadata.Metadata,
	productSlug string,
) ReleaseImageReferencesAdder {
	return ReleaseImageReferencesAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name ReleaseImageReferencesAdderClient . releaseImageReferencesAdderClient
type releaseImageReferencesAdderClient interface {
	ImageReferences(productSlug string) ([]pivnet.ImageReference, error)
	AddImageReference(productSlug string, releaseID int, imageReferenceID int) error
	CreateImageReference(config pivnet.CreateImageReferenceConfig) (pivnet.ImageReference, error)
}

type imageReferenceKey struct {
	Name      string
	ImagePath string
	Digest    string
}

func (rf ReleaseImageReferencesAdder) AddReleaseImageReferences(release pivnet.Release) error {
	productImageReferences, err := rf.pivnet.ImageReferences(rf.productSlug)
	if err != nil {
		return err
	}

	var productImageReferenceMap = map[imageReferenceKey]int{}
	for _, productImageReference := range productImageReferences {
		productImageReferenceMap[imageReferenceKey{
			productImageReference.Name,
			productImageReference.ImagePath,
			productImageReference.Digest,
		}] = productImageReference.ID
	}

	for _, imageReference := range rf.metadata.ImageReferences {
		var imageReferenceID = imageReference.ID

		if imageReferenceID == 0 {
			foundImageReferenceId := productImageReferenceMap[imageReferenceKey{
				imageReference.Name,
				imageReference.ImagePath,
				imageReference.Digest,
			}]

			if foundImageReferenceId != 0 {
				imageReferenceID = foundImageReferenceId
			} else {
				rf.logger.Info(fmt.Sprintf(
					"Creating image reference with name: %s",
					imageReference.Name,
				))

				ir, err := rf.pivnet.CreateImageReference(pivnet.CreateImageReferenceConfig{
					ProductSlug:        rf.productSlug,
					Name:               imageReference.Name,
					ImagePath:          imageReference.ImagePath,
					Digest:             imageReference.Digest,
					Description:        imageReference.Description,
					DocsURL:            imageReference.DocsURL,
					SystemRequirements: imageReference.SystemRequirements,
				})

				if err != nil {
					return err
				}

				imageReferenceID = ir.ID
			}
		}

		rf.logger.Info(fmt.Sprintf(
			"Adding image reference with ID: %d",
			imageReferenceID,
		))
		err := rf.pivnet.AddImageReference(rf.productSlug, release.ID, imageReferenceID)
		if err != nil {
			return err
		}
	}

	return nil
}
