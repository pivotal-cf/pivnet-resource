package sorter

import (
	"fmt"
	"sort"
	"time"

	"github.com/blang/semver"
	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
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

func (s Sorter) SortByLastUpdated(input []pivnet.Release) ([]pivnet.Release, error) {
	releasesMap := make(map[int64][]pivnet.Release)

	for _, release := range input {
		t, err := getMostRecentTimestampFromRelease(release)
		if err != nil {
			return nil, err
		}

		releasesMap[t] = append(releasesMap[t], release)
	}

	return sortReleasesByTimestamp(releasesMap), nil
}

func toStrings(input semver.Versions) []string {
	strings := make([]string, len(input))

	for i, v := range input {
		strings[i] = v.String()
	}

	return strings
}

func getMostRecentTimestampFromRelease(release pivnet.Release) (int64, error) {
	updatedAtTimestamp, err := time.Parse(time.RFC3339, release.UpdatedAt)
	if err != nil {
		return 0, err
	}

	mostRecentTimestamp := updatedAtTimestamp.Unix()

	if release.UserGroupsUpdatedAt != "" {
		userGroupUpdatedAtTimestamp, err := time.Parse(time.RFC3339, release.UserGroupsUpdatedAt)
		if err != nil {
			return 0, err
		}

		if mostRecentTimestamp < userGroupUpdatedAtTimestamp.Unix() {
			mostRecentTimestamp = userGroupUpdatedAtTimestamp.Unix()
		}
	}

	return mostRecentTimestamp, nil
}

func sortReleasesByTimestamp(releasesMap map[int64][]pivnet.Release) (result []pivnet.Release) {
	var keys []int64

	for key := range releasesMap {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	for _, key := range keys {
		result = append(result, releasesMap[key]...)
	}

	return result
}
