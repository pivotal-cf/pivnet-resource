package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateReleaseConfig struct {
	ProductName    string
	ProductVersion string
	ReleaseType    string
	ReleaseDate    string
	EulaSlug       string
	Description    string
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
