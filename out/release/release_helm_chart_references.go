package release

import (
	"fmt"
	pivnet "github.com/pivotal-cf/go-pivnet/v4"
	"github.com/pivotal-cf/go-pivnet/v4/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"time"
)

type ReleaseHelmChartReferencesAdder struct {
	logger      logger.Logger
	pivnet      releaseHelmChartReferencesAdderClient
	metadata    metadata.Metadata
	productSlug string
	pollFrequency time.Duration
	asyncTimeout time.Duration
}

func NewReleaseHelmChartReferencesAdder(
	logger logger.Logger,
	pivnetClient releaseHelmChartReferencesAdderClient,
	metadata metadata.Metadata,
	productSlug string,
	pollFrequency time.Duration,
	asyncTimeout time.Duration,
) ReleaseHelmChartReferencesAdder {
	return ReleaseHelmChartReferencesAdder{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
		pollFrequency: pollFrequency,
		asyncTimeout: asyncTimeout,
	}
}

//go:generate counterfeiter --fake-name ReleaseHelmChartReferencesAdderClient . releaseHelmChartReferencesAdderClient
type releaseHelmChartReferencesAdderClient interface {
	HelmChartReferences(productSlug string) ([]pivnet.HelmChartReference, error)
	AddHelmChartReference(productSlug string, releaseID int, helmChartReferenceID int) error
	CreateHelmChartReference(config pivnet.CreateHelmChartReferenceConfig) (pivnet.HelmChartReference, error)
	GetHelmChartReference(productSlug string, helmChartReferenceID int) (pivnet.HelmChartReference, error)
	DeleteHelmChartReference(productSlug string, helmChartReferenceID int) (pivnet.HelmChartReference, error)
}

type helmChartReferenceKey struct {
	Name    string
	Version string
}

func (rf ReleaseHelmChartReferencesAdder) AddReleaseHelmChartReferences(release pivnet.Release) error {
	productHelmChartReferences, err := rf.pivnet.HelmChartReferences(rf.productSlug)
	if err != nil {
		return err
	}

	var productHelmChartReferenceMap = map[helmChartReferenceKey]int{}
	for _, productHelmChartReference := range productHelmChartReferences {
		productHelmChartReferenceMap[helmChartReferenceKey{
			productHelmChartReference.Name,
			productHelmChartReference.Version,
		}] = productHelmChartReference.ID
	}

	// add references to product
	for i, helmChartReference := range rf.metadata.HelmChartReferences {
		var helmChartReferenceID = helmChartReference.ID

		if helmChartReferenceID == 0 {
			foundHelmChartReferenceId := productHelmChartReferenceMap[helmChartReferenceKey{
				helmChartReference.Name,
				helmChartReference.Version,
			}]

			if foundHelmChartReferenceId != 0 {
				helmChartReferenceID = foundHelmChartReferenceId
			} else {
				rf.logger.Info(fmt.Sprintf(
					"Creating helm chart reference with name: %s",
					helmChartReference.Name,
				))

				ir, err := rf.pivnet.CreateHelmChartReference(pivnet.CreateHelmChartReferenceConfig{
					ProductSlug:        rf.productSlug,
					Name:               helmChartReference.Name,
					Version:            helmChartReference.Version,
					Description:        helmChartReference.Description,
					DocsURL:            helmChartReference.DocsURL,
					SystemRequirements: helmChartReference.SystemRequirements,
				})

				if err != nil {
					return err
				}

				helmChartReferenceID = ir.ID
			}
			rf.metadata.HelmChartReferences[i].ID = helmChartReferenceID
		}
	}

	pollTicker := time.NewTicker(rf.pollFrequency)
	for _, helmChartReference := range rf.metadata.HelmChartReferences {
		var helmChartReferenceID = helmChartReference.ID

		rf.logger.Info(fmt.Sprintf(
			"Checking replication status of helm chart reference with name: %s",
			helmChartReference.Name,
		))
		timeoutTimer := time.NewTimer(rf.asyncTimeout)

		for {
			replicated := false
			select {
			case <-timeoutTimer.C:
				return fmt.Errorf("timed out replicating helm chart reference with name: %s", helmChartReference.Name)
			case <-pollTicker.C:
				ref, err := rf.pivnet.GetHelmChartReference(rf.productSlug, helmChartReferenceID)

				if err != nil {
					return err
				} else if ref.ReplicationStatus == pivnet.FailedToReplicate {
					return fmt.Errorf("helm chart reference with name %s failed to replicate", ref.Name)
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
	for _, helmChartReference := range rf.metadata.HelmChartReferences {
		var helmChartReferenceID = helmChartReference.ID

		rf.logger.Info(fmt.Sprintf(
			"Adding helm chart reference with ID: %d",
			helmChartReferenceID,
		))
		err := rf.pivnet.AddHelmChartReference(rf.productSlug, release.ID, helmChartReferenceID)
		if err != nil {
			return err
		}
	}

	return nil
}
