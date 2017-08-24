package uaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pivotal-cf/go-pivnet"
	"net/http"
)

type AuthResp struct {
	Token string `json:"access_token"`
}

type TokenFetcher struct {
	Endpoint     string
	RefreshToken string
}

func NewTokenFetcher(endpoint, refresh_token string) *TokenFetcher {
	return &TokenFetcher{endpoint, refresh_token}
}

func (t TokenFetcher) GetToken() (string, error) {
	httpClient := &http.Client{}
	body := pivnet.AuthBody{RefreshToken: t.RefreshToken}
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API token request body: %s", err.Error())
	}

	req, err := http.NewRequest("POST", t.Endpoint+"/api/v2/access_tokens", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to construct API token request: %s", err.Error())
	}

	resp, err := httpClient.Do(req)
	defer resp.Body.Close()

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
