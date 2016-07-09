package check_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/check"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter/filterfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp/gpfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/sorter/sorterfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Check", func() {
	var (
		fakeFilter         *filterfakes.FakeFilter
		fakePivnetClient   *gpfakes.FakeClient
		fakeExtendedClient *gpfakes.FakeExtendedClient
		fakeSorter         *sorterfakes.FakeSorter

		testLogger lager.Logger

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand

		releaseTypes    []string
		releaseTypesErr error

		allReleases      []pivnet.Release
		releasesErr      error
		filteredReleases []pivnet.Release

		releasesByReleaseTypeErr error
		releasesByVersionErr     error
		etagErr                  error
	)

	BeforeEach(func() {
		fakeFilter = &filterfakes.FakeFilter{}
		fakePivnetClient = &gpfakes.FakeClient{}
		fakeExtendedClient = &gpfakes.FakeExtendedClient{}
		fakeSorter = &sorterfakes.FakeSorter{}

		releasesByReleaseTypeErr = nil
		releasesByVersionErr = nil
		releaseTypesErr = nil
		releasesErr = nil
		etagErr = nil

		releaseTypes = []string{
			"foo release",
			"bar",
			"third release type",
		}

		allReleases = []pivnet.Release{
			{
				ID:          1,
				Version:     "A",
				ReleaseType: releaseTypes[0],
			},
			{
				ID:          2,
				Version:     "C",
				ReleaseType: releaseTypes[1],
			},
			{
				ID:          3,
				Version:     "B",
				ReleaseType: releaseTypes[2],
			},
		}

		filteredReleases = allReleases

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
			},
		}

	})

	JustBeforeEach(func() {
		fakePivnetClient.ReleaseTypesReturns(releaseTypes, releaseTypesErr)
		fakePivnetClient.ReleasesForProductSlugReturns(allReleases, releasesErr)

		fakeExtendedClient.ReleaseETagStub = func(productSlug string, releaseID int) (string, error) {
			etag := fmt.Sprintf("etag-%d", releaseID)
			return etag, etagErr
		}

		fakeFilter.ReleasesByReleaseTypeReturns(filteredReleases, releasesByReleaseTypeErr)
		fakeFilter.ReleasesByVersionReturns(filteredReleases, releasesByVersionErr)

		binaryVersion := "v0.1.2-unit-tests"

		testLogger = lagertest.NewTestLogger("check unit tests")

		checkCommand = check.NewCheckCommand(
			binaryVersion,
			testLogger,
			fakeFilter,
			fakePivnetClient,
			fakeExtendedClient,
			fakeSorter,
		)
	})

	It("returns the most recent version without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		expectedVersionWithEtag, err := versions.CombineVersionAndETag(
			allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(HaveLen(1))
		Expect(response[0].ProductVersion).To(Equal(expectedVersionWithEtag))
	})

	Context("when no releases are returned", func() {
		BeforeEach(func() {
			allReleases = []pivnet.Release{}
		})

		It("returns empty response without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(BeEmpty())
		})
	})

	Context("when there is an error getting release types", func() {
		BeforeEach(func() {
			releaseTypesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

	Context("when there is an error getting releases", func() {
		BeforeEach(func() {
			releasesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

	Context("when there is an error getting etag", func() {
		BeforeEach(func() {
			etagErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

	Describe("when a version is provided", func() {
		Context("when the version is the latest", func() {
			BeforeEach(func() {
				versionWithETag, err := versions.CombineVersionAndETag(
					allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
				)
				Expect(err).NotTo(HaveOccurred())

				checkRequest.Version = concourse.Version{
					ProductVersion: versionWithETag,
				}
			})

			It("returns the most recent version", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				versionWithETagA, err := versions.CombineVersionAndETag(
					allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(1))
				Expect(response[0].ProductVersion).To(Equal(versionWithETagA))
			})
		})

		Context("when the version is not the latest", func() {
			BeforeEach(func() {
				versionWithETag, err := versions.CombineVersionAndETag(
					allReleases[2].Version, fmt.Sprintf("etag-%d", allReleases[2].ID),
				)
				Expect(err).NotTo(HaveOccurred())

				checkRequest.Version = concourse.Version{
					ProductVersion: versionWithETag,
				}
			})

			It("returns the most recent version", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				versionWithETagC, err := versions.CombineVersionAndETag(
					allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
				)
				Expect(err).NotTo(HaveOccurred())

				versionWithETagA, err := versions.CombineVersionAndETag(
					allReleases[0].Version, fmt.Sprintf("etag-%d", allReleases[0].ID),
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(2))
				Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
				Expect(response[1].ProductVersion).To(Equal(versionWithETagA))
			})
		})
	})

	Context("when the release type is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = releaseTypes[1]

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		It("returns the most recent version with that release type", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagC, err := versions.CombineVersionAndETag(
				allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
		})

		Context("when the release type is invalid", func() {
			BeforeEach(func() {
				checkRequest.Source.ReleaseType = "not a valid release type"
			})

			It("returns an error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*release_type.*one of"))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[0]))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[1]))
				Expect(err.Error()).To(ContainSubstring(releaseTypes[2]))
			})
		})

		Context("when filtering returns an error", func() {
			BeforeEach(func() {
				releasesByReleaseTypeErr = fmt.Errorf("some release type error")
			})

			It("returns the error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(releasesByReleaseTypeErr))
			})
		})
	})

	Context("when the product version is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = releaseTypes[1]

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		BeforeEach(func() {
			checkRequest.Source.ProductVersion = "C"
		})

		It("returns the newest release with that version without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionWithETagC, err := versions.CombineVersionAndETag(
				allReleases[1].Version, fmt.Sprintf("etag-%d", allReleases[1].ID),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(versionWithETagC))
		})

		Context("when filtering returns an error", func() {
			BeforeEach(func() {
				releasesByVersionErr = fmt.Errorf("some version error")
			})

			It("returns the error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(releasesByVersionErr))
			})
		})
	})

	Context("when sorting by semver", func() {
		var (
			semverOrderedReleases []pivnet.Release
		)

		BeforeEach(func() {
			checkRequest.Source.SortBy = concourse.SortBySemver
			checkRequest.Version = concourse.Version{
				ProductVersion: "1.2.3#etag-5432",
			}

			semverOrderedReleases = []pivnet.Release{
				{
					ID:      7654,
					Version: "2.3.4",
				},
				{
					ID:      6543,
					Version: "1.2.4",
				},
				{
					ID:      5432,
					Version: "1.2.3",
				},
			}

			fakeSorter.SortBySemverReturns(semverOrderedReleases, nil)
		})

		It("returns in ascending semver order", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			versionsWithETags := make([]string, len(semverOrderedReleases))

			versionsWithETags[0], err = versions.CombineVersionAndETag(
				"1.2.4", fmt.Sprintf("etag-%d", 6543),
			)
			versionsWithETags[1], err = versions.CombineVersionAndETag(
				"2.3.4", fmt.Sprintf("etag-%d", 7654),
			)

			Expect(response).To(HaveLen(2))
			Expect(response[0].ProductVersion).To(Equal(versionsWithETags[0]))
			Expect(response[1].ProductVersion).To(Equal(versionsWithETags[1]))
			Expect(fakeSorter.SortBySemverCallCount()).To(Equal(1))
		})

		Context("when sorting by semver returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakeSorter.SortBySemverReturns(nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})
	})
})
