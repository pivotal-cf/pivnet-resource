package pivnet

import (
	"fmt"
	"net/http"
)

type ReleaseUpgradePathsService struct {
	client Client
}

type ReleaseUpgradePathsResponse struct {
	ReleaseUpgradePaths []ReleaseUpgradePath `json:"upgrade_paths,omitempty"`
}

type ReleaseUpgradePath struct {
	Release UpgradePathRelease `json:"release,omitempty" yaml:"release,omitempty"`
}

type UpgradePathRelease struct {
	ID      int    `json:"id,omitempty" yaml:"id,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

func (r ReleaseUpgradePathsService) Get(productSlug string, releaseID int) ([]ReleaseUpgradePath, error) {
	url := fmt.Sprintf(
		"/products/%s/releases/%d/upgrade_paths",
		productSlug,
		releaseID,
	)

	var response ReleaseUpgradePathsResponse
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

	return response.ReleaseUpgradePaths, nil
}
