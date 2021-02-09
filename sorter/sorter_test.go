package sorter_test

import (
	"fmt"
	"log"

	bsemver "github.com/blang/semver"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/sorter"
	"github.com/pivotal-cf/pivnet-resource/v3/sorter/sorterfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorter", func() {
	var (
		fakeSemverConverter *sorterfakes.FakeSemverConverter

		s *sorter.Sorter
	)

	BeforeEach(func() {
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger := logshim.NewLogShim(logger, logger, true)
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

		s = sorter.NewSorter(fakeLogger, fakeSemverConverter)
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

	Describe("Sort by last updated", func() {
		It("should sort releases based on most recent updated at", func() {
			pair1 := updatePair{"2019-03-12T12:23:45.430Z", "2019-04-07T06:23:13.430Z"}
			pair2 := updatePair{"2019-02-20T18:21:13.430Z","2019-11-20T08:00:13.430Z"}
			pair3 := updatePair{"2019-10-07T04:15:03.430Z", "2019-04-23T08:55:12.430Z"}

			input := releasesWithLastUpdated(pair1, pair2, pair3)

			returned, err := s.SortByLastUpdated(input)

			Expect(err).NotTo(HaveOccurred())
			Expect(lastUpdatedAtFromReleases(returned)).To(Equal([]updatePair{pair2, pair3, pair1}))
		})

		Context("when user group last updated at time is empty", func() {
			It("should sort releases based on most recent updated at", func() {
				pair1 := updatePair{"2019-03-12T12:23:45.430Z", "2019-04-07T06:23:13.430Z"}
				pair2 := updatePair{"2019-02-20T18:21:13.430Z",""}
				pair3 := updatePair{"2019-10-07T04:15:03.430Z", "2019-04-23T08:55:12.430Z"}

				input := releasesWithLastUpdated(pair1, pair2, pair3)

				returned, err := s.SortByLastUpdated(input)

				Expect(err).NotTo(HaveOccurred())
				Expect(lastUpdatedAtFromReleases(returned)).To(Equal([]updatePair{pair3, pair1, pair2}))
			})
		})

		Context("when last updated at time is the same", func() {
			It("should sort releases based on most recent updated at", func() {
				pair1 := updatePair{"2019-11-20T08:00:13.430Z", "2019-04-07T06:23:13.430Z"}
				pair2 := updatePair{"2019-02-20T18:21:13.430Z","2019-11-20T08:00:13.430Z"}
				pair3 := updatePair{"2019-10-07T04:15:03.430Z", "2019-04-23T08:55:12.430Z"}

				input := releasesWithLastUpdated(pair1, pair2, pair3)

				returned, err := s.SortByLastUpdated(input)

				Expect(err).NotTo(HaveOccurred())
				Expect(lastUpdatedAtFromReleases(returned)).To(Equal([]updatePair{pair1, pair2, pair3}))
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

func releasesWithLastUpdated(lastUpdated ...updatePair) []pivnet.Release {
	var releases []pivnet.Release
	for _, timePair := range lastUpdated {
		releases = append(releases, pivnet.Release{UpdatedAt: timePair.ReleaseUpdateAt, UserGroupsUpdatedAt: timePair.UserGroupsUpdateAt})
	}
	return releases
}

func lastUpdatedAtFromReleases(releases []pivnet.Release) []updatePair {
	var updatePairs []updatePair
	for _, release := range releases {
		updatePairs = append(updatePairs, updatePair{release.UpdatedAt, release.UserGroupsUpdatedAt})
	}

	return updatePairs
}

type updatePair struct {
	ReleaseUpdateAt    string
	UserGroupsUpdateAt string
}
