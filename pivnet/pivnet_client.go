package pivnet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

const (
	URL = "https://network.pivotal.io/api/v2"
)

type Client interface {
	ProductVersions(string) ([]string, error)
	CreateRelease(config CreateReleaseConfig) (Release, error)
	GetRelease(string, string) (Release, error)
	GetProductFiles(Release) (ProductFiles, error)
	CreateProductFile(config CreateProductFileConfig) (ProductFile, error)
	DeleteProductFile(productName string, id int) (ProductFile, error)
}

type client struct {
	url    string
	token  string
	logger logger.Logger
}

func NewClient(url, token string, logger logger.Logger) Client {
	return &client{
		url:    url,
		token:  token,
		logger: logger,
	}
}

func (c client) ProductVersions(id string) ([]string, error) {
	url := c.url + "/products/" + id + "/releases"

	var response Response
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

	var versions []string
	for _, r := range response.Releases {
		versions = append(versions, r.Version)
	}

	return versions, nil
}

func (c client) makeRequest(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) error {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf(
			"Pivnet returned status code: %d for the request - expected %d",
			resp.StatusCode,
			expectedStatusCode,
		)
	}

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return err
	}

	return nil
}
