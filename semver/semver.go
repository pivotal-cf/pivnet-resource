package semver

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/pivotal-cf/go-pivnet/logger"
)

type SemverConverter struct {
	logger logger.Logger
}

func NewSemverConverter(logger logger.Logger) *SemverConverter {
	return &SemverConverter{logger}
}

// ToValidSemver attempts to return the input as valid semver.
// If the input fails to parse as semver, it appends .0 or .0.0 to the input and retries
// If this is still not valid semver, it returns an error
func (s SemverConverter) ToValidSemver(input string) (semver.Version, error) {
	v, err := semver.Parse(input)
	if err == nil {
		return v, nil
	}

	s.logger.Info(fmt.Sprintf(
		"failed to parse semver: '%s', appending zeros and trying again",
		input,
	))
	maybeSemver := input

	segs := strings.SplitN(maybeSemver, ".", 3)
	switch len(segs) {
	case 2:
		maybeSemver += ".0"
	case 1:
		maybeSemver += ".0.0"
	}

	v, err = semver.Parse(maybeSemver)
	if err == nil {
		return v, nil
	}

	s.logger.Info(fmt.Sprintf(
		"still failed to parse semver: '%s', giving up",
		maybeSemver,
	))

	return semver.Version{}, err
}
