package sorter

import (
	"strings"

	"github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

//go:generate counterfeiter . Sorter

type Sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
}

type sorter struct {
}

func NewSorter() Sorter {
	return &sorter{}
}

func (sorter) SortBySemver(input []pivnet.Release) ([]pivnet.Release, error) {
	versions := make([]string, len(input))
	versionsToReleases := make(map[string]pivnet.Release)

	for i, release := range input {
		appended := appendZerosIfNecessary(release.Version)
		versionsToReleases[appended] = release
		versions[i] = appended
	}

	semverVersions, err := toSemverVersions(versions)
	if err != nil {
		return nil, err
	}
	semver.Sort(semverVersions)

	sortedStrings := toStrings(semverVersions)

	// we know we want to order highest-first
	reversedStrings := reverse(sortedStrings)

	sortedReleases := make([]pivnet.Release, len(input))

	for i, v := range reversedStrings {
		sortedReleases[i] = versionsToReleases[v]
	}

	return sortedReleases, nil
}

func toSemverVersions(input []string) (semver.Versions, error) {
	versions := make([]semver.Version, len(input))

	for i, s := range input {
		v, err := semver.Parse(s)
		if err != nil {
			return nil, err
		}
		versions[i] = v
	}

	return versions, nil
}

func appendZerosIfNecessary(input string) string {
	probablySemver := input

	segs := strings.SplitN(probablySemver, ".", 3)
	switch len(segs) {
	case 2:
		probablySemver += ".0"
	case 1:
		probablySemver += ".0.0"
	}

	return probablySemver
}

func toStrings(input semver.Versions) []string {
	strings := make([]string, len(input))

	for i, v := range input {
		strings[i] = v.String()
	}

	return strings
}

func reverse(input []string) []string {
	count := len(input)
	reversed := make([]string, count)
	for i, s := range input {
		reversed[count-1-i] = s
	}
	return reversed
}
