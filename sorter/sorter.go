package sorter

import (
	"strings"

	"github.com/blang/semver"
)

//go:generate counterfeiter . Sorter

type Sorter interface {
	SortBySemver([]string) ([]string, error)
}

type sorter struct {
}

func NewSorter() Sorter {
	return &sorter{}
}

func (sorter) SortBySemver(input []string) ([]string, error) {
	versions, err := toSemverVersions(input)
	if err != nil {
		return nil, err
	}
	semver.Sort(versions)

	sortedStrings := toStrings(versions)

	return reverse(sortedStrings), nil
}

func toSemverVersions(input []string) (semver.Versions, error) {
	versions := make([]semver.Version, len(input))

	for i, s := range input {
		probablySemver := s
		segs := strings.SplitN(probablySemver, ".", 3)
		switch len(segs) {
		case 2:
			probablySemver += ".0"
		case 1:
			probablySemver += ".0.0"
		}

		v, err := semver.Parse(probablySemver)
		if err != nil {
			return nil, err
		}
		versions[i] = v
	}

	return versions, nil
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
