package sorter

import (
	"fmt"
	"strings"

	"github.com/pivotal-golang/lager"

	"github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

//go:generate counterfeiter . Sorter

type Sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
}

type sorter struct {
	logger lager.Logger
}

func NewSorter(logger lager.Logger) Sorter {
	return &sorter{
		logger: logger,
	}
}

// SortBySemver returns the provided releases, ordered by semantic versioning,
// in descending order i.e. [4.2.3, 1.2.1, 1.2.0]
// If a version cannot be parsed as semantic versioning, this is logged to stdout
// and that release is not returned. No error is returned in this case.
// Therefore the number of returned releases may be fewer than the number of
// provided releases.
func (s sorter) SortBySemver(input []pivnet.Release) ([]pivnet.Release, error) {
	var versions []string
	versionsToReleases := make(map[string]pivnet.Release)

	for _, release := range input {
		appended := s.toValidSemver(release.Version)
		if appended != "" {
			versionsToReleases[appended] = release
			versions = append(versions, appended)
		}
	}

	semverVersions := s.toSemverVersions(versions)

	semver.Sort(semverVersions)

	sortedStrings := toStrings(semverVersions)

	sortedReleases := make([]pivnet.Release, len(sortedStrings))

	// reverse
	count := len(sortedStrings)
	for i, v := range sortedStrings {
		sortedReleases[count-i-1] = versionsToReleases[v]
	}

	return sortedReleases, nil
}

// toSemverVersions expects all versions are valid semver
func (s sorter) toSemverVersions(input []string) semver.Versions {
	var versions []semver.Version

	for _, str := range input {
		v, err := semver.Parse(str)
		if err != nil {
			s.logger.Info(fmt.Sprintf("failed to parse semver: '%s' - should be valid by this point\n", str))
		} else {
			versions = append(versions, v)
		}
	}

	return versions
}

// toValidSemver attempts to return the input as valid semver.
// It appends .0 or .0.0 to the input
// If this is still not valid semver, it gives up and returns empty string
func (s sorter) toValidSemver(input string) string {
	_, err := semver.Parse(input)
	if err == nil {
		return input
	}

	s.logger.Info(fmt.Sprintf("failed to parse semver: '%s', appending zeros and trying again\n", input))
	maybeSemver := input

	segs := strings.SplitN(maybeSemver, ".", 3)
	switch len(segs) {
	case 2:
		maybeSemver += ".0"
	case 1:
		maybeSemver += ".0.0"
	}

	_, err = semver.Parse(maybeSemver)
	if err == nil {
		return maybeSemver
	}

	s.logger.Info(fmt.Sprintf("still failed to parse semver: '%s', giving up\n", maybeSemver))

	return ""
}

func toStrings(input semver.Versions) []string {
	strings := make([]string, len(input))

	for i, v := range input {
		strings[i] = v.String()
	}

	return strings
}
