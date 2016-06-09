package release

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

type ReleaseFinalizer struct {
	pivnet          updateClient
	metadataFetcher fetcher
	sourcesDir      string
	params          concourse.OutParams
	productSlug     string
}

func NewFinalizer(pivnetClient updateClient, metadataFetcher fetcher, params concourse.OutParams, sourcesDir, productSlug string) ReleaseFinalizer {
	return ReleaseFinalizer{
		pivnet:          pivnetClient,
		metadataFetcher: metadataFetcher,
		params:          params,
		sourcesDir:      sourcesDir,
		productSlug:     productSlug,
	}
}

//go:generate counterfeiter --fake-name UpdateClient . updateClient
type updateClient interface {
	UpdateRelease(string, pivnet.Release) (pivnet.Release, error)
	ReleaseETag(string, pivnet.Release) (string, error)
	AddUserGroup(productSlug string, releaseID int, userGroupID int) error
}

//go:generate counterfeiter --fake-name Fetcher . fetcher
type fetcher interface {
	Fetch(yamlKey, dir, file string) string
}

func (rf ReleaseFinalizer) Finalize(release pivnet.Release) (concourse.OutResponse, error) {
	availability := rf.metadataFetcher.Fetch("Availability", rf.sourcesDir, rf.params.AvailabilityFile)
	if availability != "Admins Only" {
		releaseUpdate := pivnet.Release{
			ID:           release.ID,
			Availability: availability,
		}

		var err error
		release, err = rf.pivnet.UpdateRelease(rf.productSlug, releaseUpdate)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		if availability == "Selected User Groups Only" {
			userGroupIDs := strings.Split(
				rf.metadataFetcher.Fetch("UserGroupIDs", rf.sourcesDir, rf.params.UserGroupIDsFile),
				",",
			)

			for _, userGroupIDString := range userGroupIDs {
				userGroupID, err := strconv.Atoi(userGroupIDString)
				if err != nil {
					panic(err)
				}

				err = rf.pivnet.AddUserGroup(rf.productSlug, release.ID, userGroupID)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	releaseETag, err := rf.pivnet.ReleaseETag(rf.productSlug, release)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	outputVersion, err := versions.CombineVersionAndETag(release.Version, releaseETag)
	if err != nil {
		return concourse.OutResponse{}, err // this will never return an error
	}

	metadata := []concourse.Metadata{
		{Name: "version", Value: release.Version},
		{Name: "release_type", Value: release.ReleaseType},
		{Name: "release_date", Value: release.ReleaseDate},
		{Name: "description", Value: release.Description},
		{Name: "release_notes_url", Value: release.ReleaseNotesURL},
		{Name: "availability", Value: release.Availability},
		{Name: "controlled", Value: fmt.Sprintf("%t", release.Controlled)},
		{Name: "eccn", Value: release.ECCN},
		{Name: "license_exception", Value: release.LicenseException},
		{Name: "end_of_support_date", Value: release.EndOfSupportDate},
		{Name: "end_of_guidance_date", Value: release.EndOfGuidanceDate},
		{Name: "end_of_availability_date", Value: release.EndOfAvailabilityDate},
	}
	if release.EULA != nil {
		metadata = append(metadata, concourse.Metadata{Name: "eula_slug", Value: release.EULA.Slug})
	}

	return concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: outputVersion,
		},
		Metadata: metadata,
	}, nil
}
