package pivnet

import (
	"fmt"
	"net/http"
	"strings"
)

func (c client) EULAs() ([]EULA, error) {
	url := fmt.Sprintf(
		"%s/eulas",
		c.url,
	)

	var response EULAsResponse
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

	return response.EULAs, nil
}

func (c client) AcceptEULA(productSlug string, releaseID int) error {
	url := fmt.Sprintf(
		"%s/products/%s/releases/%d/eula_acceptance",
		c.url,
		productSlug,
		releaseID,
	)

	var response EULAAcceptanceResponse
	err := c.makeRequest(
		"POST",
		url,
		http.StatusOK,
		strings.NewReader(`{}`),
		&response,
	)
	if err != nil {
		return err
	}

	return nil
}
