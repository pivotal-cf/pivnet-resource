package gp

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/download"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"io"
	"net/http"
)

type Client struct {
	client pivnet.Client
}

func NewClient(token pivnet.AccessTokenService, config pivnet.ClientConfig, logger logger.Logger) *Client {
	return &Client{
		client: pivnet.NewClient(token, config, logger),
	}
}

func (c Client) GetFederationToken(productSlug string) (pivnet.FederationToken, error) {
	return c.client.FederationToken.GenerateFederationToken(productSlug)
}

func (c Client) ReleaseTypes() ([]pivnet.ReleaseType, error) {
	return c.client.ReleaseTypes.Get()
}

func (c Client) S3PrefixForProductSlug(productSlug string) (string, error) {
	product, err := c.client.Products.Get(productSlug)
	if err != nil {
		return "", err
	}

	return product.S3Directory.Path, nil
}

func (c Client) ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error) {
	return c.client.Releases.List(productSlug)
}

func (c Client) GetRelease(productSlug string, version string) (pivnet.Release, error) {
	releases, err := c.client.Releases.List(productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}

	var foundRelease pivnet.Release
	for _, r := range releases {
		if r.Version == version {
			foundRelease = r
			break
		}
	}

	if foundRelease.Version != version {
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

func (c Client) DeleteRelease(productSlug string, release pivnet.Release) error {
	return c.client.Releases.Delete(productSlug, release)
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

func (c Client) FindProductForSlug(slug string) (pivnet.Product, error) {
	return c.client.Products.Get(slug)
}

func (c Client) ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error) {
	return c.client.ProductFiles.ListForRelease(productSlug, releaseID)
}

func (c Client) ProductFiles(productSlug string) ([]pivnet.ProductFile, error) {
	return c.client.ProductFiles.List(productSlug)
}

func (c Client) ProductFile(productSlug string, productFileID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Get(productSlug, productFileID)
}

func (c Client) ProductFileForRelease(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.GetForRelease(productSlug, releaseID, productFileID)
}

func (c Client) DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Delete(productSlug, releaseID)
}

func (c Client) CreateProductFile(config pivnet.CreateProductFileConfig) (pivnet.ProductFile, error) {
	return c.client.ProductFiles.Create(config)
}

func (c Client) AddProductFile(productSlug string, releaseID int, productFileID int) error {
	return c.client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
}

func (c Client) CreateFileGroup(config pivnet.CreateFileGroupConfig) (pivnet.FileGroup, error) {
	return c.client.FileGroups.Create(config)
}

func (c Client) AddToFileGroup(productSlug string, fileGroupID int, productFileID int) error {
	return c.client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
}

func (c Client) AddFileGroup(productSlug string, releaseID int, fileGroupID int) error {
	return c.client.FileGroups.AddToRelease(productSlug, releaseID, fileGroupID)
}

func (c Client) DownloadProductFile(writer *download.FileInfo, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error {
	return c.client.ProductFiles.DownloadForRelease(writer, productSlug, releaseID, productFileID, progressWriter)
}

func (c Client) FileGroupsForRelease(productSlug string, releaseID int) ([]pivnet.FileGroup, error) {
	return c.client.FileGroups.ListForRelease(productSlug, releaseID)
}

func (c Client) ArtifactReferences(productSlug string) ([]pivnet.ArtifactReference, error) {
	return c.client.ArtifactReferences.List(productSlug)
}

func (c Client) ArtifactReferencesForRelease(productSlug string, releaseID int) ([]pivnet.ArtifactReference, error) {
	return c.client.ArtifactReferences.ListForRelease(productSlug, releaseID)
}

func (c Client) CreateArtifactReference(config pivnet.CreateArtifactReferenceConfig) (pivnet.ArtifactReference, error) {
	return c.client.ArtifactReferences.Create(config)
}

func (c Client) GetArtifactReference(productSlug string, artifactReferenceID int) (pivnet.ArtifactReference, error) {
	return c.client.ArtifactReferences.Get(productSlug, artifactReferenceID)
}

func (c Client) AddArtifactReference(productSlug string, releaseID int, artifactReferenceID int) error {
	return c.client.ArtifactReferences.AddToRelease(productSlug, releaseID, artifactReferenceID)
}

func (c Client) DeleteArtifactReference(productSlug string, artifactReferenceID int) (pivnet.ArtifactReference, error) {
	return c.client.ArtifactReferences.Delete(productSlug, artifactReferenceID)
}

func (c Client) ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error) {
	return c.client.ReleaseDependencies.List(productSlug, releaseID)
}

func (c Client) AddReleaseDependency(productSlug string, releaseID int, dependentReleaseID int) error {
	return c.client.ReleaseDependencies.Add(productSlug, releaseID, dependentReleaseID)
}

func (c Client) DependencySpecifiers(productSlug string, releaseID int) ([]pivnet.DependencySpecifier, error) {
	return c.client.DependencySpecifiers.List(productSlug, releaseID)
}

func (c Client) CreateDependencySpecifier(productSlug string, releaseID int, dependentProductSlug string, specifier string) (pivnet.DependencySpecifier, error) {
	return c.client.DependencySpecifiers.Create(productSlug, releaseID, dependentProductSlug, specifier)
}

func (c Client) ReleaseUpgradePaths(productSlug string, releaseID int) ([]pivnet.ReleaseUpgradePath, error) {
	return c.client.ReleaseUpgradePaths.Get(productSlug, releaseID)
}

func (c Client) UpgradePathSpecifiers(productSlug string, releaseID int) ([]pivnet.UpgradePathSpecifier, error) {
	return c.client.UpgradePathSpecifiers.List(productSlug, releaseID)
}

func (c Client) CreateUpgradePathSpecifier(productSlug string, releaseID int, specifier string) (pivnet.UpgradePathSpecifier, error) {
	return c.client.UpgradePathSpecifiers.Create(productSlug, releaseID, specifier)
}

func (c Client) AddReleaseUpgradePath(productSlug string, releaseID int, previousReleaseID int) error {
	return c.client.ReleaseUpgradePaths.Add(productSlug, releaseID, previousReleaseID)
}

func (c Client) CreateRequest(method string, url string, body io.Reader) (*http.Request, error) {
	return c.client.CreateRequest(method, url, body)
}
