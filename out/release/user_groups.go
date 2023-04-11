package release

import (
	"fmt"
	"strconv"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
)

type UserGroupsUpdater struct {
	logger      logger.Logger
	pivnet      userGroupsUpdaterClient
	metadata    metadata.Metadata
	productSlug string
}

func NewUserGroupsUpdater(
	logger logger.Logger,
	pivnetClient userGroupsUpdaterClient,
	metadata metadata.Metadata,
	productSlug string,
) UserGroupsUpdater {
	return UserGroupsUpdater{
		logger:      logger,
		pivnet:      pivnetClient,
		metadata:    metadata,
		productSlug: productSlug,
	}
}

//counterfeiter:generate --fake-name UserGroupsUpdaterClient . userGroupsUpdaterClient
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

		rf.logger.Info(fmt.Sprintf(
			"Updating availability to: '%s'",
			availability,
		))

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

				rf.logger.Info(fmt.Sprintf(
					"Adding user group with ID: %d",
					userGroupID,
				))
				err = rf.pivnet.AddUserGroup(rf.productSlug, release.ID, userGroupID)
				if err != nil {
					return pivnet.Release{}, err
				}
			}
		}
	}

	return release, nil
}
