package versions

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
)

const (
	etagDelimiter = "#"
)

func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i], nil
		}
	}

	return versions[:1], nil
}

func Reverse(versions []string) ([]string, error) {
	var reversed []string
	for i := len(versions) - 1; i >= 0; i-- {
		reversed = append(reversed, versions[i])
	}

	return reversed, nil
}

func SplitIntoVersionAndETag(versionWithETag string) (string, string, error) {
	split := strings.Split(versionWithETag, etagDelimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("Invalid version and Etag: %s", versionWithETag)
	}
	return split[0], split[1], nil
}

func CombineVersionAndETag(version string, etag string) (string, error) {
	if etag == "" {
		return version, nil
	}

	return fmt.Sprintf("%s%s%s", version, etagDelimiter, etag), nil
}

//go:generate counterfeiter --fake-name FakeExtendedClient . extendedClient
type extendedClient interface {
	ReleaseETag(productSlug string, releaseID int) (string, error)
}

// ProductVersions adds the release ETags to the release versions
func ProductVersions(
	c extendedClient,
	productSlug string,
	releases []pivnet.Release,
) ([]string, error) {
	var versions []string
	for _, r := range releases {
		etag, err := c.ReleaseETag(productSlug, r.ID)
		if err != nil {
			return nil, err
		}

		version := fmt.Sprintf("%s#%s", r.Version, etag)
		versions = append(versions, version)
	}

	return versions, nil
}
