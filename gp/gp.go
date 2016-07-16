package gp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

//go:generate counterfeiter . Client
type Client interface {
	ReleaseTypes() ([]string, error)
	GetRelease(productSlug string, productVersion string) (pivnet.Release, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	UpdateRelease(productslug string, release pivnet.Release) (pivnet.Release, error)
	CreateRelease(pivnet.CreateReleaseConfig) (pivnet.Release, error)

	AcceptEULA(productSlug string, releaseID int) error
	EULAs() ([]pivnet.EULA, error)

	FindProductForSlug(slug string) (pivnet.Product, error)
	CreateProductFile(pivnet.CreateProductFileConfig) (pivnet.ProductFile, error)
	AddProductFile(productSlug string, releaseID int, productFileID int) error
	GetProductFiles(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	GetProductFile(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error)

	AddUserGroup(productSlug string, releaseID int, userGroupID int) error

	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)

	MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error)
}

//go:generate counterfeiter . ExtendedClient
type ExtendedClient interface {
	ReleaseETag(productSlug string, releaseID int) (string, error)
	ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error)
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

func (c client) GetRelease(productSlug string, productVersion string) (pivnet.Release, error) {
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

func (c client) UpdateRelease(productSlug string, release pivnet.Release) (pivnet.Release, error) {
	return c.client.Releases.Update(productSlug, release)
}

func (c client) CreateRelease(config pivnet.CreateReleaseConfig) (pivnet.Release, error) {
	return c.client.Releases.Create(config)
}

func (c client) AddUserGroup(productSlug string, releaseID int, userGroupID int) error {
	return c.client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
}

func (c client) AcceptEULA(productSlug string, releaseID int) error {
	return c.client.EULA.Accept(productSlug, releaseID)
}

func (c client) EULAs() ([]pivnet.EULA, error) {
	return c.client.EULA.List()
}

func (c client) GetProductFiles(productSlug string, releaseID int) ([]pivnet.ProductFile, error) {
	return c.client.ProductFiles.ListForRelease(productSlug, releaseID)
}

func (c client) GetProductFile(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.GetForRelease(productSlug, releaseID, productFileID)
}

func (c client) FindProductForSlug(slug string) (pivnet.Product, error) {
	return c.client.Products.Get(slug)
}

func (c client) CreateProductFile(config pivnet.CreateProductFileConfig) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Create(config)
}

func (c client) AddProductFile(productSlug string, releaseID int, productFileID int) error {
	return c.client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
}

func (c client) ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error) {
	return c.client.ReleaseDependencies.List(productSlug, releaseID)
}

func (c client) MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error) {
	return c.client.MakeRequest(method, url, expectedResponseCode, body, data)
}

func (c extendedClient) ReleaseETag(productSlug string, releaseID int) (string, error) {
	return c.client.ReleaseETag(productSlug, releaseID)
}

func (c extendedClient) ProductVersions(productSlug string, releases []pivnet.Release) ([]string, error) {
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
