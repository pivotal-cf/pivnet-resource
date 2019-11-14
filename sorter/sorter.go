package sorter

import (
	"fmt"

	"github.com/blang/semver"
	pivnet "github.com/pivotal-cf/go-pivnet/v3"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
)

//go:generate counterfeiter --fake-name FakeSemverConverter . semverConverter
type semverConverter interface {
	ToValidSemver(string) (semver.Version, error)
}

type Sorter struct {
	logger          logger.Logger
	semverConverter semverConverter
}

func NewSorter(logger logger.Logger, semverConverter semverConverter) *Sorter {
	return &Sorter{
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
func (s Sorter) SortBySemver(input []pivnet.Release) ([]pivnet.Release, error) {
	var versions []semver.Version
	versionsToReleases := make(map[string]pivnet.Release)

	for _, release := range input {
		asSemver, err := s.semverConverter.ToValidSemver(release.Version)
		if err != nil {
			s.logger.Info(fmt.Sprintf(
				"failed to parse release version as semver: '%s'",
				release.Version,
			))
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
