package uaa

import (
	"net/http"
	"github.com/pivotal-cf/go-pivnet"
	"encoding/json"
	"bytes"
	"fmt"
)

type AuthResp struct {
	Token string `json: "token"`
}

type TokenFetcher struct {
	Endpoint string
	Username string
	Password string
}

func NewTokenFetcher(endpoint, username, password string) *TokenFetcher {
	return &TokenFetcher{endpoint, username, password}
}

func (t TokenFetcher) GetToken() (string, error) {
	httpClient := &http.Client{}
	body := pivnet.AuthBody{Username: t.Username, Password: t.Password}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API token request body: %s", err.Error())
	}

	req, err := http.NewRequest("POST", t.Endpoint + "/api/v2/authentication", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to construct API token request: %s", err.Error())
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API token request failed: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch API token - received status %v", resp.StatusCode)
	}

	var response AuthResp
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode API token response: %s", err.Error())
	}

	return response.Token, nil
}
