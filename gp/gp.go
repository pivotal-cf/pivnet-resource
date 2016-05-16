package gp

import (
	"io"
	"net/http"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

//go:generate counterfeiter . Client

type Client interface {
	ReleaseTypes() ([]string, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error)
}

//go:generate counterfeiter . ExtendedClient

type ExtendedClient interface {
	ReleaseETag(productSlug string, releaseID int) (string, error)
}

type client struct {
	client pivnet.Client
}

func NewClient(config pivnet.ClientConfig, logger logger.Logger) Client {
	return &client{
		client: pivnet.NewClient(config, logger),
	}
}

type extendedClient struct {
	client extension.ExtendedClient
}

func NewExtendedClient(c Client, logger logger.Logger) ExtendedClient {
	return &extendedClient{
		client: extension.NewExtendedClient(c, logger),
	}
}

func (c client) ReleaseTypes() ([]string, error) {
	return c.client.ReleaseTypes.Get()
}

func (c client) ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error) {
	return c.client.Releases.List(productSlug)
}

func (c client) MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error) {
	return c.client.MakeRequest(method, url, expectedResponseCode, body, data)
}

func (c extendedClient) ReleaseETag(productSlug string, releaseID int) (string, error) {
	return c.client.ReleaseETag(productSlug, releaseID)
}
