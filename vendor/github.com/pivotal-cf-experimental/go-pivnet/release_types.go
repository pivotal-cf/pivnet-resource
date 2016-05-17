package pivnet

import (
	"fmt"
	"net/http"
)

type ReleaseTypesService struct {
	client Client
}

type ReleaseTypesResponse struct {
	ReleaseTypes []string `json:"release_types" yaml:"release_types"`
}

func (r ReleaseTypesService) Get() ([]string, error) {
	url := fmt.Sprintf("/releases/release_types")

	var response ReleaseTypesResponse
	_, err := r.client.MakeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.ReleaseTypes, nil
}
