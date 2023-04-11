package release

import (
	"fmt"
	"time"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
)

type ReleaseArtifactReferencesAdder struct {
	logger        logger.Logger
	pivnet        releaseArtifactReferencesAdderClient
	metadata      metadata.Metadata
	productSlug   string
	pollFrequency time.Duration
	asyncTimeout  time.Duration
}

func NewReleaseArtifactReferencesAdder(
	logger logger.Logger,
	pivnetClient releaseArtifactReferencesAdderClient,
	metadata metadata.Metadata,
	productSlug string,
	pollFrequency time.Duration,
	asyncTimeout time.Duration,
) ReleaseArtifactReferencesAdder {
	return ReleaseArtifactReferencesAdder{
		logger:        logger,
		pivnet:        pivnetClient,
		metadata:      metadata,
		productSlug:   productSlug,
		pollFrequency: pollFrequency,
		asyncTimeout:  asyncTimeout,
	}
}

//counterfeiter:generate --fake-name ReleaseArtifactReferencesAdderClient . releaseArtifactReferencesAdderClient
type releaseArtifactReferencesAdderClient interface {
	ArtifactReferences(productSlug string) ([]pivnet.ArtifactReference, error)
	ArtifactReferencesForDigest(productSlug string, digest string) ([]pivnet.ArtifactReference, error)
	AddArtifactReference(productSlug string, releaseID int, artifactReferenceID int) error
	CreateArtifactReference(config pivnet.CreateArtifactReferenceConfig) (pivnet.ArtifactReference, error)
	GetArtifactReference(productSlug string, artifactReferenceID int) (pivnet.ArtifactReference, error)
	DeleteArtifactReference(productSlug string, artifactReferenceID int) (pivnet.ArtifactReference, error)
}

type artifactReferenceKey struct {
	Name         string
	ArtifactPath string
	Digest       string
}

func (rf ReleaseArtifactReferencesAdder) AddReleaseArtifactReferences(release pivnet.Release) error {
	// add references to product
	for i, artifactReference := range rf.metadata.ArtifactReferences {
		var artifactReferenceID = artifactReference.ID

		if artifactReferenceID == 0 {
			foundArtifactReferences, err := rf.pivnet.ArtifactReferencesForDigest(rf.productSlug, artifactReference.Digest)
			if err != nil {
				return err
			}

			if len(foundArtifactReferences) > 0 {
				artifactReferenceID = foundArtifactReferences[0].ID
			} else {
				rf.logger.Info(fmt.Sprintf(
					"Creating artifact reference with name: %s",
					artifactReference.Name,
				))

				ir, err := rf.pivnet.CreateArtifactReference(pivnet.CreateArtifactReferenceConfig{
					ProductSlug:        rf.productSlug,
					Name:               artifactReference.Name,
					ArtifactPath:       artifactReference.ArtifactPath,
					Digest:             artifactReference.Digest,
					Description:        artifactReference.Description,
					DocsURL:            artifactReference.DocsURL,
					SystemRequirements: artifactReference.SystemRequirements,
				})

				if err != nil {
					return err
				}

				artifactReferenceID = ir.ID
			}
			rf.metadata.ArtifactReferences[i].ID = artifactReferenceID
		}
	}

	// wait for references to replicate
	pollTicker := time.NewTicker(rf.pollFrequency)
	for _, artifactReference := range rf.metadata.ArtifactReferences {
		var artifactReferenceID = artifactReference.ID

		rf.logger.Info(fmt.Sprintf(
			"Checking replication status of artifact reference with name: %s",
			artifactReference.Name,
		))
		timeoutTimer := time.NewTimer(rf.asyncTimeout)

		for {
			replicated := false
			select {
			case <-timeoutTimer.C:
				return fmt.Errorf("timed out replicating artifact reference with name: %s", artifactReference.Name)
			case <-pollTicker.C:
				ref, err := rf.pivnet.GetArtifactReference(rf.productSlug, artifactReferenceID)

				if err != nil {
					return err
				} else if ref.ReplicationStatus == pivnet.FailedToReplicate {
					return fmt.Errorf("artifact reference with name %s failed to replicate", ref.Name)
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
	for _, artifactReference := range rf.metadata.ArtifactReferences {
		var artifactReferenceID = artifactReference.ID

		rf.logger.Info(fmt.Sprintf(
			"Adding artifact reference with ID: %d",
			artifactReferenceID,
		))
		err := rf.pivnet.AddArtifactReference(rf.productSlug, release.ID, artifactReferenceID)
		if err != nil {
			return err
		}
	}

	return nil
}
