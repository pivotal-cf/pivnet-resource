package release

import (
	"fmt"

	pivnet "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logger"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/versions"
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

//go:generate counterfeiter --fake-name FinalizerClient . finalizerClient
type finalizerClient interface {
	GetRelease(productSlug string, releaseVersion string) (pivnet.Release, error)
}

func (rf ReleaseFinalizer) Finalize(productSlug string, releaseVersion string) (concourse.OutResponse, error) {
	newRelease, err := rf.pivnet.GetRelease(productSlug, releaseVersion)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	outputVersion, err := versions.CombineVersionAndFingerprint(newRelease.Version, newRelease.SoftwareFilesUpdatedAt)
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
