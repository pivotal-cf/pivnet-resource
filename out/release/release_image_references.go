package release

import (
	"fmt"
	pivnet "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"time"
)

type ReleaseImageReferencesAdder struct {
	logger      logger.Logger
	pivnet      releaseImageReferencesAdderClient
	metadata    metadata.Metadata
	productSlug string
	pollFrequency time.Duration
	asyncTimeout time.Duration
}

func NewReleaseImageReferencesAdder(
	logger logger.Logger,
	pivnetClient releaseImageReferencesAdderClient,
	metadata metadata.Metadata,
	productSlug string,
	pollFrequency time.Duration,
	asyncTimeout time.Duration,
) ReleaseImageReferencesAdder {
	return ReleaseImageReferencesAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
		pollFrequency: pollFrequency,
		asyncTimeout: asyncTimeout,
	}
}

//go:generate counterfeiter --fake-name ReleaseImageReferencesAdderClient . releaseImageReferencesAdderClient
type releaseImageReferencesAdderClient interface {
	ImageReferences(productSlug string) ([]pivnet.ImageReference, error)
	AddImageReference(productSlug string, releaseID int, imageReferenceID int) error
	CreateImageReference(config pivnet.CreateImageReferenceConfig) (pivnet.ImageReference, error)
	GetImageReference(productSlug string, imageReferenceID int) (pivnet.ImageReference, error)
	DeleteImageReference(productSlug string, imageReferenceID int) (pivnet.ImageReference, error)
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

	// add references to product
	for i, imageReference := range rf.metadata.ImageReferences {
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
			rf.metadata.ImageReferences[i].ID = imageReferenceID
		}
	}

	// wait for references to replicate
	pollTicker := time.NewTicker(rf.pollFrequency)
	for _, imageReference := range rf.metadata.ImageReferences {
		var imageReferenceID = imageReference.ID

		rf.logger.Info(fmt.Sprintf(
			"Checking replication status of image reference with name: %s",
			imageReference.Name,
		))
		timeoutTimer := time.NewTimer(rf.asyncTimeout)

		for {
			replicated := false
			select {
			case <-timeoutTimer.C:
				return fmt.Errorf("timed out replicating image reference with name: %s", imageReference.Name)
			case <-pollTicker.C:
				ref, err := rf.pivnet.GetImageReference(rf.productSlug, imageReferenceID)

				if err != nil {
					return err
				} else if ref.ReplicationStatus == pivnet.FailedToReplicate {
					return fmt.Errorf("image reference with name %s failed to replicate", ref.Name)
				} else if ref.ReplicationStatus == pivnet.Complete {
					replicated = true
				}
			}
			if replicated {
				break
			}
		}
	}

	// add references to release
	for _, imageReference := range rf.metadata.ImageReferences {
		var imageReferenceID = imageReference.ID

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
