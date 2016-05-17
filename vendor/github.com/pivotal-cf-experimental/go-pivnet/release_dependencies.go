package pivnet

import (
	"fmt"
	"net/http"
)

type ReleaseDependenciesService struct {
	client Client
}

type ReleaseDependenciesResponse struct {
	ReleaseDependencies []ReleaseDependency `json:"dependencies,omitempty"`
}

type ReleaseDependency struct {
	Release DependentRelease `json:"release,omitempty" yaml:"release,omitempty"`
}

type DependentRelease struct {
	ID      int     `json:"id,omitempty" yaml:"id,omitempty"`
	Version string  `json:"version,omitempty" yaml:"version,omitempty"`
	Product Product `json:"product,omitempty" yaml:"product,omitempty"`
}

func (r ReleaseDependenciesService) Get(productSlug string, releaseID int) ([]ReleaseDependency, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/dependencies",
		productSlug,
		releaseID,
	)

	var response ReleaseDependenciesResponse
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

	return response.ReleaseDependencies, nil
}
