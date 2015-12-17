package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

	response, err := c.makeRequest("GET", releasesURL, nil)
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
	releasesURL := c.url + "/products/" + productName + "/releases"

	response, err := c.makeRequest("GET", releasesURL, nil)
	if err != nil {
		panic(err)
	}

	var matchingRelease Release

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

func (c client) makeRequest(requestType string, url string, body io.Reader) (Response, error) {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return Response{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("Pivnet returned status code: %d for the request", resp.StatusCode)
	}

	response := Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return Response{}, err
	}

	return response, err
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

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", releasesURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return Release{}, nil
}

type createReleaseBody struct {
	Release Release `json:"release"`
}
