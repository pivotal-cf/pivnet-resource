package pivnet

import (
	"fmt"
	"net/http"
)

func (c client) ReleaseDependencies(productSlug string, releaseID int) ([]ReleaseDependency, error) {
	url := fmt.Sprintf(
		"%s/products/%s/releases/%d/dependencies",
		c.url,
		productSlug,
		releaseID,
	)

	var response ReleaseDependenciesResponse
	err := c.makeRequest(
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
