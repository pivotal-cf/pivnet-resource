package pivnet

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	Version string `json:"version"`
}

type Client interface {
	ProductVersions(string) ([]string, error)
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
		panic(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token: %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	response := Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	var versions []string
	for _, r := range response.Releases {
		versions = append(versions, r.Version)
	}

	return versions, nil
}
