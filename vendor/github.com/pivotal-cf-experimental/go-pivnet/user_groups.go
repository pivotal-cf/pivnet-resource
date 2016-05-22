package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserGroupsService struct {
	client Client
}

type addRemoveUserGroupBody struct {
	UserGroup UserGroup `json:"user_group"`
}

type createUserGroupBody struct {
	UserGroup createUserGroup `json:"user_group"`
}

type updateUserGroupBody struct {
	UserGroup updateUserGroup `json:"user_group"`
}

type UserGroupsResponse struct {
	UserGroups []UserGroup `json:"user_groups,omitempty"`
}

type UpdateUserGroupResponse struct {
	UserGroup UserGroup `json:"user_group,omitempty"`
}

type UserGroup struct {
	ID          int      `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Members     []string `json:"members,omitempty" yaml:"members,omitempty"`
}

type createUserGroup struct {
	ID          int      `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members"` // do not omit empty to satisfy pivnet
}

type updateUserGroup struct {
	ID          int      `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members,omitempty"`
}

func (u UserGroupsService) List() ([]UserGroup, error) {
	url := "/user_groups"

	var response UserGroupsResponse
	_, err := u.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.UserGroups, nil
}

func (u UserGroupsService) ListForRelease(productSlug string, releaseID int) ([]UserGroup, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/user_groups",
		productSlug,
		releaseID,
	)

	var response UserGroupsResponse
	_, err := u.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.UserGroups, nil
}

func (u UserGroupsService) AddToRelease(productSlug string, releaseID int, userGroupID int) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/add_user_group",
		productSlug,
		releaseID,
	)

	body := addRemoveUserGroupBody{
		UserGroup: UserGroup{
			ID: userGroupID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	_, err = u.client.MakeRequest(
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

func (u UserGroupsService) RemoveFromRelease(productSlug string, releaseID int, userGroupID int) error {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/remove_user_group",
		productSlug,
		releaseID,
	)

	body := addRemoveUserGroupBody{
		UserGroup: UserGroup{
			ID: userGroupID,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	_, err = u.client.MakeRequest(
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

func (u UserGroupsService) Get(userGroupID int) (UserGroup, error) {
	url := fmt.Sprintf("/user_groups/%d", userGroupID)

	var response UserGroup
	_, err := u.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return UserGroup{}, err
	}

	return response, nil
}

func (u UserGroupsService) Create(name string, description string, members []string) (UserGroup, error) {
	url := "/user_groups"

	if members == nil {
		members = []string{}
	}

	createBody := createUserGroupBody{
		createUserGroup{
			Name:        name,
			Description: description,
			Members:     members,
		},
	}

	b, err := json.Marshal(createBody)
	if err != nil {
		return UserGroup{}, err
	}

	body := bytes.NewReader(b)

	var response UserGroup
	_, err = u.client.MakeRequest(
		"POST",
		url,
		http.StatusCreated,
		body,
		&response,
	)
	if err != nil {
		return UserGroup{}, err
	}

	return response, nil
}

func (u UserGroupsService) Update(userGroup UserGroup) (UserGroup, error) {
	url := fmt.Sprintf("/user_groups/%d", userGroup.ID)

	createBody := updateUserGroupBody{
		updateUserGroup{
			Name:        userGroup.Name,
			Description: userGroup.Description,
		},
	}

	b, err := json.Marshal(createBody)
	if err != nil {
		return UserGroup{}, err
	}

	body := bytes.NewReader(b)

	var response UpdateUserGroupResponse
	_, err = u.client.MakeRequest(
		"PATCH",
		url,
		http.StatusOK,
		body,
		&response,
	)
	if err != nil {
		return UserGroup{}, err
	}

	return response.UserGroup, nil
}

func (r UserGroupsService) Delete(userGroupID int) error {
	url := fmt.Sprintf("/user_groups/%d", userGroupID)

	_, err := r.client.MakeRequest(
		"DELETE",
		url,
		http.StatusNoContent,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
