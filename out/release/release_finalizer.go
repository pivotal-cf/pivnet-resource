package release

import (
	"fmt"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
	"github.com/pivotal-cf/pivnet-resource/v3/versions"
)

type ReleaseFinalizer struct {
	logger      logger.Logger
	pivnet      finalizerClient
	metadata    metadata.Metadata
	params      concourse.OutParams
	sourcesDir  string
	productSlug string
}

func NewFinalizer(
	pivnetClient finalizerClient,
	logger logger.Logger,
	params concourse.OutParams,
	metadata metadata.Metadata,
	sourcesDir,
	productSlug string,
) ReleaseFinalizer {
	return ReleaseFinalizer{
		pivnet:      pivnetClient,
		logger:      logger,
		params:      params,
		metadata:    metadata,
		sourcesDir:  sourcesDir,
		productSlug: productSlug,
	}
}

//counterfeiter:generate --fake-name FinalizerClient . finalizerClient
type finalizerClient interface {
	GetRelease(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	FileGroupsForRelease(productSlug string, releaseID int) ([]pivnet.FileGroup, error)
}

func (rf ReleaseFinalizer) Finalize(productSlug string, releaseVersion string) (concourse.OutResponse, error) {
	newRelease, err := rf.pivnet.GetRelease(productSlug, releaseVersion)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	// Compute fingerprint from product files so version only changes when files change (TNZ-22056)
	releaseProductFiles, err := rf.pivnet.ProductFilesForRelease(productSlug, newRelease.ID)
	if err != nil {
		return concourse.OutResponse{}, err
	}
	fileGroups, err := rf.pivnet.FileGroupsForRelease(productSlug, newRelease.ID)
	if err != nil {
		return concourse.OutResponse{}, err
	}
	seen := make(map[int]bool)
	var fileMetadata []versions.FileMetadata
	for _, pf := range releaseProductFiles {
		if !seen[pf.ID] {
			seen[pf.ID] = true
			fileMetadata = append(fileMetadata, versions.FileMetadata{ID: pf.ID, ReleasedAt: pf.ReleasedAt})
		}
	}
	for _, fg := range fileGroups {
		for _, pf := range fg.ProductFiles {
			if !seen[pf.ID] {
				seen[pf.ID] = true
				fileMetadata = append(fileMetadata, versions.FileMetadata{ID: pf.ID, ReleasedAt: pf.ReleasedAt})
			}
		}
	}
	fingerprint := versions.FingerprintFromFileMetadata(fileMetadata)

	outputVersion, err := versions.CombineVersionAndFingerprint(newRelease.Version, fingerprint)
	if err != nil {
		return concourse.OutResponse{}, err // this will never return an error
	}

	metadata := []concourse.Metadata{
		{Name: "version", Value: newRelease.Version},
		{Name: "release_type", Value: string(newRelease.ReleaseType)},
		{Name: "release_date", Value: newRelease.ReleaseDate},
		{Name: "description", Value: newRelease.Description},
		{Name: "release_notes_url", Value: newRelease.ReleaseNotesURL},
		{Name: "availability", Value: newRelease.Availability},
		{Name: "controlled", Value: fmt.Sprintf("%t", newRelease.Controlled)},
		{Name: "eccn", Value: newRelease.ECCN},
		{Name: "license_exception", Value: newRelease.LicenseException},
		{Name: "end_of_support_date", Value: newRelease.EndOfSupportDate},
		{Name: "end_of_guidance_date", Value: newRelease.EndOfGuidanceDate},
		{Name: "end_of_availability_date", Value: newRelease.EndOfAvailabilityDate},
	}
	if newRelease.EULA != nil {
		metadata = append(
			metadata,
			concourse.Metadata{Name: "eula_slug", Value: newRelease.EULA.Slug})
	}

	return concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: outputVersion,
		},
		Metadata: metadata,
	}, nil
}
