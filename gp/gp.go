package gp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

type CombinedClient struct {
	*Client
	*ExtendedClient
}

type Client struct {
	client pivnet.Client
}

func NewClient(config pivnet.ClientConfig, logger logger.Logger) *Client {
	return &Client{
		client: pivnet.NewClient(config, logger),
	}
}

type ExtendedClient struct {
	client extension.ExtendedClient
}

func NewExtendedClient(c Client, logger logger.Logger) *ExtendedClient {
	return &ExtendedClient{
		client: extension.NewExtendedClient(c, logger),
	}
}

func (c Client) ReleaseTypes() ([]string, error) {
	return c.client.ReleaseTypes.Get()
}

func (c Client) ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error) {
	return c.client.Releases.List(productSlug)
}

func (c Client) GetRelease(productSlug string, productVersion string) (pivnet.Release, error) {
	releases, err := c.client.Releases.List(productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}

	var foundRelease pivnet.Release
	for _, r := range releases {
		if r.Version == productVersion {
			foundRelease = r
			break
		}
	}

	if foundRelease.Version != productVersion {
		return pivnet.Release{}, fmt.Errorf("release not found")
	}

	release, err := c.client.Releases.Get(productSlug, foundRelease.ID)
	if err != nil {
		return pivnet.Release{}, err
	}
	return release, nil
}

func (c Client) UpdateRelease(productSlug string, release pivnet.Release) (pivnet.Release, error) {
	return c.client.Releases.Update(productSlug, release)
}

func (c Client) CreateRelease(config pivnet.CreateReleaseConfig) (pivnet.Release, error) {
	return c.client.Releases.Create(config)
}

func (c Client) AddUserGroup(productSlug string, releaseID int, userGroupID int) error {
	return c.client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
}

func (c Client) UserGroups(productSlug string, releaseID int) ([]pivnet.UserGroup, error) {
	return c.client.UserGroups.ListForRelease(productSlug, releaseID)
}

func (c Client) AcceptEULA(productSlug string, releaseID int) error {
	return c.client.EULA.Accept(productSlug, releaseID)
}

func (c Client) EULAs() ([]pivnet.EULA, error) {
	return c.client.EULA.List()
}

func (c Client) GetProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error) {
	return c.client.ProductFiles.ListForRelease(productSlug, releaseID)
}

func (c Client) GetProductFiles(productSlug string) ([]pivnet.ProductFile, error) {
	return c.client.ProductFiles.List(productSlug)
}

func (c Client) GetProductFile(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.GetForRelease(productSlug, releaseID, productFileID)
}

func (c Client) DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Delete(productSlug, releaseID)
}

func (c Client) FindProductForSlug(slug string) (pivnet.Product, error) {
	return c.client.Products.Get(slug)
}

func (c Client) CreateProductFile(config pivnet.CreateProductFileConfig) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Create(config)
}

func (c Client) AddProductFile(productSlug string, releaseID int, productFileID int) error {
	return c.client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
}

func (c Client) ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error) {
	return c.client.ReleaseDependencies.List(productSlug, releaseID)
}

func (c Client) MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error) {
	return c.client.MakeRequest(method, url, expectedResponseCode, body, data)
}

func (c ExtendedClient) ReleaseETag(productSlug string, releaseID int) (string, error) {
	return c.client.ReleaseETag(productSlug, releaseID)
}

func (c ExtendedClient) ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error) {
	var versions []string
	for _, r := range releases {
		etag, err := c.client.ReleaseETag(productSlug, r.ID)
		if err != nil {
			return nil, err
		}
		version := fmt.Sprintf("%s#%s", r.Version, etag)
		versions = append(versions, version)
	}

	return versions, nil
}
