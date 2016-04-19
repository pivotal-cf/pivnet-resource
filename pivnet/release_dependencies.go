package pivnet

import (
	"fmt"
	"net/http"
)

func (c client) ReleaseDependencies(productID int, releaseID int) ([]ReleaseDependency, error) {
	url := fmt.Sprintf(
		"%s/products/%d/releases/%d/dependencies",
		c.url,
		productID,
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
