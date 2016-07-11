package sorter

import (
	"log"

	"github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

//go:generate counterfeiter . SemverConverter
type SemverConverter interface {
	ToValidSemver(string) (semver.Version, error)
}

//go:generate counterfeiter . Sorter
type Sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
}

type sorter struct {
	logger          *log.Logger
	semverConverter SemverConverter
}

func NewSorter(logger *log.Logger, semverConverter SemverConverter) Sorter {
	return &sorter{
		logger:          logger,
		semverConverter: semverConverter,
	}
}

// SortBySemver returns the provided releases, ordered by semantic versioning,
// in descending order i.e. [4.2.3, 1.2.1, 1.2.0]
// If a version cannot be parsed as semantic versioning, this is logged to stdout
// and that release is not returned. No error is returned in this case.
// Therefore the number of returned releases may be fewer than the number of
// provided releases.
func (s sorter) SortBySemver(input []pivnet.Release) ([]pivnet.Release, error) {
	var versions []semver.Version
	versionsToReleases := make(map[string]pivnet.Release)

	for _, release := range input {
		asSemver, err := s.semverConverter.ToValidSemver(release.Version)
		if err != nil {
			s.logger.Printf("failed to parse release version as semver: '%s'", release.Version)
			continue
		}

		versionsToReleases[asSemver.String()] = release
		versions = append(versions, asSemver)
	}

	semver.Sort(versions)

	sortedStrings := toStrings(versions)

	sortedReleases := make([]pivnet.Release, len(sortedStrings))

	count := len(sortedStrings)
	for i, v := range sortedStrings {
		// perform reversal so we return descending, not ascending
		sortedReleases[count-i-1] = versionsToReleases[v]
	}

	return sortedReleases, nil
}

func toStrings(input semver.Versions) []string {
	strings := make([]string, len(input))

	for i, v := range input {
		strings[i] = v.String()
	}

	return strings
}
