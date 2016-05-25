package pivnet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

const (
	DefaultHost = "https://network.pivotal.io"
	apiVersion  = "/api/v2"
)

type pivnetErr struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type ErrPivnetOther struct {
	ResponseCode int      `json:"response_code" yaml:"response_code"`
	Message      string   `json:"message" yaml:"message"`
	Errors       []string `json:"errors" yaml:"errors"`
}

func (e ErrPivnetOther) Error() string {
	return fmt.Sprintf(
		"%d - %s. Errors: %v",
		e.ResponseCode,
		e.Message,
		strings.Join(e.Errors, ","),
	)
}

type ErrUnauthorized struct {
	ResponseCode int    `json:"response_code" yaml:"response_code"`
	Message      string `json:"message" yaml:"message"`
}

func (e ErrUnauthorized) Error() string {
	return e.Message
}

func newErrUnauthorized(message string) ErrUnauthorized {
	return ErrUnauthorized{
		ResponseCode: http.StatusUnauthorized,
		Message:      message,
	}
}

type ErrNotFound struct {
	ResponseCode int    `json:"response_code" yaml:"response_code"`
	Message      string `json:"message" yaml:"message"`
}

func (e ErrNotFound) Error() string {
	return e.Message
}

func newErrNotFound(message string) ErrNotFound {
	return ErrNotFound{
		ResponseCode: http.StatusNotFound,
		Message:      message,
	}
}

type Client struct {
	baseURL   string
	token     string
	userAgent string
	logger    logger.Logger

	Auth                *AuthService
	EULA                *EULAsService
	ProductFiles        *ProductFilesService
	FileGroups          *FileGroupsService
	Releases            *ReleasesService
	Products            *ProductsService
	UserGroups          *UserGroupsService
	ReleaseDependencies *ReleaseDependenciesService
	ReleaseTypes        *ReleaseTypesService
	ReleaseUpgradePaths *ReleaseUpgradePathsService
}

type ClientConfig struct {
	Host      string
	Token     string
	UserAgent string
}

func NewClient(config ClientConfig, logger logger.Logger) Client {
	baseURL := fmt.Sprintf("%s%s", config.Host, apiVersion)

	client := Client{
		baseURL:   baseURL,
		token:     config.Token,
		userAgent: config.UserAgent,
		logger:    logger,
	}

	client.Auth = &AuthService{client: client}
	client.EULA = &EULAsService{client: client}
	client.ProductFiles = &ProductFilesService{client: client}
	client.FileGroups = &FileGroupsService{client: client}
	client.Releases = &ReleasesService{client: client, l: logger}
	client.Products = &ProductsService{client: client, l: logger}
	client.UserGroups = &UserGroupsService{client: client}
	client.ReleaseDependencies = &ReleaseDependenciesService{client: client}
	client.ReleaseTypes = &ReleaseTypesService{client: client}
	client.ReleaseUpgradePaths = &ReleaseUpgradePathsService{client: client}

	return client
}

func (c Client) makeRequestWithHTTPResponse(
	requestType string,
	url string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) (*http.Response, error) {
	req, err := http.NewRequest(requestType, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Add("User-Agent", c.userAgent)

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("Making request", logger.Data{"request": string(reqBytes)})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	c.logger.Debug("Response status code", logger.Data{"status code": resp.StatusCode})

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(b) > 0 {
		c.logger.Debug("Response body", logger.Data{"response body": string(b)})
	}

	if resp.StatusCode != expectedStatusCode {
		var pErr pivnetErr
		err = json.Unmarshal(b, &pErr)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, newErrUnauthorized(pErr.Message)
		case http.StatusNotFound:
			return nil, newErrNotFound(pErr.Message)
		}

		return nil, ErrPivnetOther{
			ResponseCode: resp.StatusCode,
			Message:      pErr.Message,
			Errors:       pErr.Errors,
		}
	}

	if len(b) > 0 && data != nil {
		err = json.Unmarshal(b, data)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c Client) MakeRequest(
	requestType string,
	endpoint string,
	expectedStatusCode int,
	body io.Reader,
	data interface{},
) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = u.Path + endpoint

	resp, err := c.makeRequestWithHTTPResponse(
		requestType,
		u.String(),
		expectedStatusCode,
		body,
		data,
	)
	return resp, err
}
