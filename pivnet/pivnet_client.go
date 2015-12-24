package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
}

type Client interface {
	ProductVersions(string) ([]string, error)
	CreateRelease(config CreateReleaseConfig) (Release, error)
	GetRelease(string, string) (Release, error)
	GetProductFiles(Release) (ProductFiles, error)
}

type client struct {
	url   string
	token string
}

func NewClient(url, token string) Client {
	return &client{
		url:   url,
		token: token,
	}
}

func (c client) ProductVersions(id string) ([]string, error) {
	releasesURL := c.url + "/products/" + id + "/releases"

	var response Response
	err := c.makeRequest(
		"GET",
		releasesURL,
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

	releasesURL := c.url + "/products/" + productName + "/releases"

	var response Response
	err := c.makeRequest(
		"GET",
		releasesURL,
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
			return Release{}, fmt.Errorf("The requested version: %s - could not be found", version)
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
		return fmt.Errorf("Pivnet returned status code: %d for the request - expected %d", resp.StatusCode, expectedStatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (c client) CreateRelease(config CreateReleaseConfig) (Release, error) {
	releasesURL := c.url + "/products/" + config.ProductName + "/releases"

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
		},
	}

	if config.ReleaseDate == "" {
		body.Release.ReleaseDate = time.Now().Format("2006-01-02")
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response CreateReleaseResponse
	err = c.makeRequest(
		"POST",
		releasesURL,
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
