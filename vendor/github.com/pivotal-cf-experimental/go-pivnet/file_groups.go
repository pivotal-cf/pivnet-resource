package pivnet

import (
	"fmt"
	"net/http"
)

type FileGroupsService struct {
	client Client
}

type FileGroup struct {
	ID           int              `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string           `json:"name,omitempty" yaml:"name,omitempty"`
	Product      FileGroupProduct `json:"product,omitempty" yaml:"product,omitempty"`
	ProductFiles []ProductFile    `json:"product_files,omitempty" yaml:"product_files,omitempty"`
}

type FileGroupProduct struct {
	ID   int    `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type FileGroupsResponse struct {
	FileGroups []FileGroup `json:"file_groups,omitempty"`
}

func (e FileGroupsService) List(productSlug string) ([]FileGroup, error) {
	url := fmt.Sprintf("/products/%s/file_groups", productSlug)

	var response FileGroupsResponse
	_, err := e.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.FileGroups, nil
}

func (p FileGroupsService) Get(productSlug string, fileGroupID int) (FileGroup, error) {
	url := fmt.Sprintf("/products/%s/file_groups/%d",
		productSlug,
		fileGroupID,
	)

	var response FileGroup
	_, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return FileGroup{}, err
	}

	return response, nil
}

func (p FileGroupsService) Delete(productSlug string, id int) (FileGroup, error) {
	url := fmt.Sprintf(
		"/products/%s/file_groups/%d",
		productSlug,
		id,
	)

	var response FileGroup
	_, err := p.client.MakeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return FileGroup{}, err
	}

	return response, nil
}

func (p FileGroupsService) ListForRelease(productSlug string, releaseID int) ([]FileGroup, error) {
	url := fmt.Sprintf("/products/%s/releases/%d/file_groups",
		productSlug,
		releaseID,
	)

	var response FileGroupsResponse
	_, err := p.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return []FileGroup{}, err
	}

	return response.FileGroups, nil
}
