package sorter_test

import (
	"fmt"
	"io/ioutil"
	"log"

	bsemver "github.com/blang/semver"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sorter"
	"github.com/pivotal-cf-experimental/pivnet-resource/sorter/sorterfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorter", func() {
	var (
		fakeSemverConverter *sorterfakes.FakeSemverConverter

		s *sorter.Sorter
	)

	BeforeEach(func() {
		testLogger := log.New(ioutil.Discard, "it doesn't matter", 0)
		fakeSemverConverter = &sorterfakes.FakeSemverConverter{}

		fakeSemverConverter.ToValidSemverStub = func(input string) (bsemver.Version, error) {
			switch input {
			case "not-semver":
				return bsemver.Version{}, fmt.Errorf("bad parse")
			case "1":
				return bsemver.Version{Major: 1}, nil
			case "1.0.0":
				return bsemver.Version{Major: 1}, nil
			case "2.0.0":
				return bsemver.Version{Major: 2}, nil
			case "2.1.0":
				return bsemver.Version{Major: 2, Minor: 1}, nil
			case "2.1":
				return bsemver.Version{Major: 2, Minor: 1}, nil
			case "2.4.1-edge.11":
				return bsemver.Version{
					Major: 2,
					Minor: 4,
					Patch: 1,
					Pre: []bsemver.PRVersion{
						{VersionStr: "edge"},
						{VersionNum: 11, IsNum: true},
					},
				}, nil
			case "2.4.1-edge.12":
				return bsemver.Version{
					Major: 2,
					Minor: 4,
					Patch: 1,
					Pre: []bsemver.PRVersion{
						{VersionStr: "edge"},
						{VersionNum: 12, IsNum: true},
					},
				}, nil
			case "2.4.1":
				return bsemver.Version{Major: 2, Minor: 4, Patch: 1}, nil
			default:
				panic(fmt.Sprintf("unrecognized input: %s", input))
			}
		}

		s = sorter.NewSorter(testLogger, fakeSemverConverter)
	})

	Describe("SortBySemver", func() {
		It("sorts descending", func() {
			input := releasesWithVersions(
				"1.0.0", "2.4.1", "2.0.0", "2.4.1-edge.12", "2.4.1-edge.11",
			)

			returned, err := s.SortBySemver(input)
			Expect(err).NotTo(HaveOccurred())

			Expect(versionsFromReleases(returned)).To(Equal(
				[]string{"2.4.1", "2.4.1-edge.12", "2.4.1-edge.11", "2.0.0", "1.0.0"}))
		})

		Context("when parsing a version as semver fails", func() {
			It("ignores that value", func() {
				input := releasesWithVersions(
					"1.0.0", "2.4.1", "not-semver",
				)

				returned, err := s.SortBySemver(input)
				Expect(err).NotTo(HaveOccurred())

				Expect(versionsFromReleases(returned)).To(Equal(
					[]string{"2.4.1", "1.0.0"}))
			})
		})

		Context("when the versions have fewer than 3 components", func() {
			It("orders as if the zeros were present", func() {
				input := releasesWithVersions(
					"1", "2.4.1", "2.1",
				)

				returned, err := s.SortBySemver(input)
				Expect(err).NotTo(HaveOccurred())

				Expect(versionsFromReleases(returned)).To(Equal(
					[]string{"2.4.1", "2.1", "1"}))
			})
		})
	})
})

func releasesWithVersions(versions ...string) []pivnet.Release {
	var releases []pivnet.Release
	for _, v := range versions {
		releases = append(releases, pivnet.Release{Version: v})
	}
	return releases
}

func versionsFromReleases(releases []pivnet.Release) []string {
	var versions []string
	for _, release := range releases {
		versions = append(versions, release.Version)
	}
	return versions
}
