package pivnet

import (
	"fmt"
	"net/http"
)

func (c client) ReleaseTypes() ([]string, error) {
	url := fmt.Sprintf(
		"%s/releases/release_types",
		c.url,
	)

	var response ReleaseTypesResponse
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

	return response.ReleaseTypes, nil
}
