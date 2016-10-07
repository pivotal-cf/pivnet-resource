package release

import (
	"strconv"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type UserGroupsUpdater struct {
	pivnet      userGroupsUpdaterClient
	metadata    metadata.Metadata
	productSlug string
}

func NewUserGroupsUpdater(
	pivnetClient userGroupsUpdaterClient,
	metadata metadata.Metadata,
	productSlug string,
) UserGroupsUpdater {
	return UserGroupsUpdater{
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//go:generate counterfeiter --fake-name UserGroupsUpdaterClient . userGroupsUpdaterClient
type userGroupsUpdaterClient interface {
	UpdateRelease(productSlug string, release pivnet.Release) (pivnet.Release, error)
	AddUserGroup(productSlug string, releaseID int, userGroupID int) error
}

func (rf UserGroupsUpdater) UpdateUserGroups(release pivnet.Release) (pivnet.Release, error) {
	availability := rf.metadata.Release.Availability

	if availability != "Admins Only" {
		releaseUpdate := pivnet.Release{
			ID:           release.ID,
			Availability: availability,
		}

		var err error
		release, err = rf.pivnet.UpdateRelease(rf.productSlug, releaseUpdate)
		if err != nil {
			return pivnet.Release{}, err
		}

		if availability == "Selected User Groups Only" {
			userGroupIDs := rf.metadata.Release.UserGroupIDs

			for _, userGroupIDString := range userGroupIDs {
				userGroupID, err := strconv.Atoi(userGroupIDString)
				if err != nil {
					return pivnet.Release{}, err
				}

				err = rf.pivnet.AddUserGroup(rf.productSlug, release.ID, userGroupID)
				if err != nil {
					return pivnet.Release{}, err
				}
			}
		}
	}

	return release, nil
}
