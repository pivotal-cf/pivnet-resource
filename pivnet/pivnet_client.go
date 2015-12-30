package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

const (
	URL = "https://network.pivotal.io/api/v2"
)

type CreateReleaseConfig struct {
	ProductName    string
	ProductVersion string
	ReleaseType    string
	ReleaseDate    string
	EulaSlug       string
	Description    string
}

type CreateProductFileConfig struct {
	ProductName  string
	FileVersion  string
	AWSObjectKey string
	Name         string
}

type Client interface {
	ProductVersions(string) ([]string, error)
	CreateRelease(config CreateReleaseConfig) (Release, error)
	GetRelease(string, string) (Release, error)
	GetProductFiles(Release) (ProductFiles, error)
	CreateProductFile(config CreateProductFileConfig) (ProductFile, error)
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

func (c client) GetRelease(productName, version string) (Release, error) {
	var matchingRelease Release

	url := c.url + "/products/" + productName + "/releases"

	var response Response
	err := c.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return Release{}, err
	}

	for i, r := range response.Releases {
		if r.Version == version {
			matchingRelease = r
			break
		}

		if i == len(response.Releases)-1 {
			return Release{}, fmt.Errorf(
				"The requested version: %s - could not be found", version)
		}
	}

	return matchingRelease, nil
}

func (c client) GetProductFiles(release Release) (ProductFiles, error) {
	productFiles := ProductFiles{}

	err := c.makeRequest(
		"GET",
		release.Links.ProductFiles["href"],
		http.StatusOK,
		nil,
		&productFiles,
	)
	if err != nil {
		return ProductFiles{}, err
	}

	return productFiles, nil
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

func (c client) CreateRelease(config CreateReleaseConfig) (Release, error) {
	url := c.url + "/products/" + config.ProductName + "/releases"

	body := createReleaseBody{
		Release: Release{
			Availability: "Admins Only",
			Eula: Eula{
				Slug: config.EulaSlug,
			},
			OSSCompliant: "confirm",
			ReleaseDate:  config.ReleaseDate,
			ReleaseType:  config.ReleaseType,
			Version:      config.ProductVersion,
			Description:  config.Description,
		},
	}

	if config.ReleaseDate == "" {
		body.Release.ReleaseDate = time.Now().Format("2006-01-02")
		c.logger.Debugf(
			"no release date found - defaulting to %s\n", body.Release.ReleaseDate)
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response CreateReleaseResponse
	err = c.makeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
		&response,
	)
	if err != nil {
		return Release{}, err
	}

	return response.Release, nil
}

type createReleaseBody struct {
	Release Release `json:"release"`
}

func (c client) CreateProductFile(config CreateProductFileConfig) (ProductFile, error) {
	url := c.url + "/products/" + config.ProductName + "/product_files"

	body := createProductFileBody{
		ProductFile: ProductFile{
			MD5:          "not-supported-yet",
			FileType:     "Software",
			FileVersion:  config.FileVersion,
			AWSObjectKey: config.AWSObjectKey,
			Name:         config.Name,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response CreateProductFileResponse
	err = c.makeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}

type createProductFileBody struct {
	ProductFile ProductFile `json:"product_file"`
}
