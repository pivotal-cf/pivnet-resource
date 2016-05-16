package pivnet_filter

import "github.com/pivotal-cf-experimental/go-pivnet"

//go:generate counterfeiter . Filter

type Filter interface {
	ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error)
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
}

type filter struct {
}

func NewFilter() Filter {
	return &filter{}
}

func (f filter) ReleasesByReleaseType(releases []pivnet.Release, releaseType string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		if release.ReleaseType == releaseType {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

func (f filter) ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		if release.Version == version {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}
