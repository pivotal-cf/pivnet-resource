package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}

type Client interface {
	ProductVersions(string) ([]string, error)
	CreateRelease(config CreateReleaseConfig) (Release, error)
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
	req, err := http.NewRequest("GET", releasesURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pivnet returned status code: %d for the request", resp.StatusCode)
	}

	response := Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, r := range response.Releases {
		versions = append(versions, r.Version)
	}

	return versions, nil
}

func (c client) CreateRelease(config CreateReleaseConfig) (Release, error) {
	releasesURL := c.url + "/products/" + config.ProductName + "/releases"

	body := createReleaseBody{
		Release: Release{
			Availability: "Admins Only",
			Eula: Eula{
				Slug: "pivotal_software_eula",
			},
			OSSCompliant: "confirm",
			ReleaseDate:  "2015-12-18",
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
