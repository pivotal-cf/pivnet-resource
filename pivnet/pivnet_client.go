package pivnet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

const (
	URL = "https://network.pivotal.io/api/v2"
)

type Client interface {
	ProductVersions(string) ([]string, error)
	CreateRelease(config CreateReleaseConfig) (Release, error)
	GetRelease(string, string) (Release, error)
	UpdateRelease(string, Release) (Release, error)
	GetProductFiles(Release) (ProductFiles, error)
	AcceptEULA(productSlug string, releaseID int) error
	CreateProductFile(config CreateProductFileConfig) (ProductFile, error)
	DeleteProductFile(productSlug string, id int) (ProductFile, error)
	AddProductFile(productID int, releaseID int, productFileID int) error
	FindProductForSlug(slug string) (Product, error)
	AddUserGroup(productSlug string, releaseID int, userGroupID int) error
}

type client struct {
	url       string
	token     string
	userAgent string
	logger    logger.Logger
}

type NewClientConfig struct {
	URL       string
	Token     string
	UserAgent string
}

func NewClient(config NewClientConfig, logger logger.Logger) Client {
	return &client{
		url:       config.URL,
		token:     config.Token,
		userAgent: config.UserAgent,
		logger:    logger,
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

func (c client) AcceptEULA(productSlug string, releaseID int) error {
	url := fmt.Sprintf("%s/products/%s/releases/%d/eula_acceptance", c.url,
		productSlug, releaseID)

	var response EulaResponse
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

func (c client) makeRequest(requestType string, url string, expectedStatusCode int, body io.Reader, data interface{}) error {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Add("User-Agent", c.userAgent)

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		c.logger.Debugf("Error dumping request: %+v\n", err)
		return err
	}

	c.logger.Debugf("Making request: %s\n", string(reqBytes))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Debugf("Error making request: %+v\n", err)
		return err
	}
	defer resp.Body.Close()

	c.logger.Debugf("Response status code: %d\n", resp.StatusCode)
	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf(
			"Pivnet returned status code: %d for the request - expected %d",
			resp.StatusCode,
			expectedStatusCode,
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(b) > 0 {
		c.logger.Debugf("Response body: %s\n", string(b))
		err = json.Unmarshal(b, data)
		if err != nil {
			return err
		}
	}

	return nil
}
