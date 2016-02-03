package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type addUserGroupBody struct {
	UserGroup UserGroup `json:"user_group"`
}

func (c client) AddUserGroup(
	productSlug string,
	releaseID int,
	userGroupID int,
) error {
	url := fmt.Sprintf(
		"%s/products/%s/releases/%d/add_user_group",
		c.url,
		productSlug,
		releaseID,
	)

	body := addUserGroupBody{
		UserGroup: UserGroup{
			ID: userGroupID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	err = c.makeRequest(
		"PATCH",
		url,
		http.StatusNoContent,
		bytes.NewReader(b),
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
