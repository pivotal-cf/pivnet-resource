package check_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/check"
	"github.com/pivotal-cf/pivnet-resource/v3/check/checkfakes"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var (
		fakeLogger       logger.Logger
		fakeFilter       *checkfakes.FakeFilter
		fakePivnetClient *checkfakes.FakePivnetClient
		fakeSorter       *checkfakes.FakeSorter

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand

		expectedVersions []string // product version only, same order as releaseVersions from check

		releaseTypes    []pivnet.ReleaseType
		releaseTypesErr error

		allReleases      []pivnet.Release
		releasesErr      error
		filteredReleases []pivnet.Release

		releasesByReleaseTypeErr error
		releasesByVersionErr     error

		tempDir     string
		logFilePath string
	)

	BeforeEach(func() {
		fakeFilter = &checkfakes.FakeFilter{}
		fakePivnetClient = &checkfakes.FakePivnetClient{}
		fakeSorter = &checkfakes.FakeSorter{}

		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		releasesByReleaseTypeErr = nil
		releasesByVersionErr = nil
		releaseTypesErr = nil
		releasesErr = nil

		releaseTypes = []pivnet.ReleaseType{
			pivnet.ReleaseType("foo release"),
			pivnet.ReleaseType("bar"),
			pivnet.ReleaseType("third release type"),
		}

		allReleases = []pivnet.Release{
			{
				ID:                     1,
				Version:                "1.2.3",
				ReleaseType:            releaseTypes[0],
				SoftwareFilesUpdatedAt: "time1",
			},
			{
				ID:                     2,
				Version:                "2.3.4",
				ReleaseType:            releaseTypes[1],
				SoftwareFilesUpdatedAt: "time2",
			},
			{
				ID:                     3,
				Version:                "1.2.4",
				ReleaseType:            releaseTypes[2],
				SoftwareFilesUpdatedAt: "time3",
			},
		}

		expectedVersions = make([]string, len(allReleases))
		for i, r := range allReleases {
			expectedVersions[i] = r.Version
		}

		filteredReleases = allReleases

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
			},
		}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "pivnet-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		fakePivnetClient.ReleaseTypesReturns(releaseTypes, releaseTypesErr)
		fakePivnetClient.ReleasesForProductSlugReturns(allReleases, releasesErr)

		fakeFilter.ReleasesByReleaseTypeReturns(filteredReleases, releasesByReleaseTypeErr)
		fakeFilter.ReleasesByVersionReturns(filteredReleases, releasesByVersionErr)

		binaryVersion := "v0.1.2-unit-tests"

		checkCommand = check.NewCheckCommand(
			fakeLogger,
			binaryVersion,
			fakeFilter,
			fakePivnetClient,
			fakeSorter,
			logFilePath,
		)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns the most recent version without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(HaveLen(1))
		Expect(response[0].ProductVersion).To(Equal(expectedVersions[0]))
	})

	Context("when no releases are returned", func() {
		BeforeEach(func() {
			allReleases = []pivnet.Release{}
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("cannot find specified release"))
		})
	})

	Context("when log files already exist", func() {
		var (
			otherFilePath1 string
			otherFilePath2 string
		)

		BeforeEach(func() {
			otherFilePath1 = filepath.Join(tempDir, "pivnet-resource-check.log1")
			err := ioutil.WriteFile(otherFilePath1, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			otherFilePath2 = filepath.Join(tempDir, "pivnet-resource-check.log2")
			err = ioutil.WriteFile(otherFilePath2, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes the other log files", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(otherFilePath1)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(otherFilePath2)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(logFilePath)
			Expect(err).NotTo(HaveOccurred())
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

	Describe("when a version is provided", func() {
		Context("when the version is the latest", func() {
			BeforeEach(func() {
				checkRequest.Version = concourse.Version{
					ProductVersion: expectedVersions[0], // 1.2.3
				}
			})

			It("returns the most recent version", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(1))
				Expect(response[0].ProductVersion).To(Equal(expectedVersions[0]))
			})
		})

		Context("when the version is not the latest", func() {
			BeforeEach(func() {
				checkRequest.Version = concourse.Version{
					ProductVersion: expectedVersions[2], // 1.2.4
				}
			})

			It("returns the most recent versions, including the version specified", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				// vs order is 1.2.3, 2.3.4, 1.2.4; Since(1.2.4) then Reverse => 1.2.4, 2.3.4, 1.2.3
				Expect(response).To(HaveLen(3))
				Expect(response[0].ProductVersion).To(Equal("1.2.4"))
				Expect(response[1].ProductVersion).To(Equal("2.3.4"))
				Expect(response[2].ProductVersion).To(Equal("1.2.3"))
			})
		})
	})

	Context("when multiple releases have the same product version", func() {
		BeforeEach(func() {
			allReleases = []pivnet.Release{
				{ID: 1, Version: "1.2.3", ReleaseType: releaseTypes[0], SoftwareFilesUpdatedAt: "time1"},
				{ID: 2, Version: "1.2.3", ReleaseType: releaseTypes[1], SoftwareFilesUpdatedAt: "time2"},
				{ID: 3, Version: "2.0.0", ReleaseType: releaseTypes[2], SoftwareFilesUpdatedAt: "time3"},
			}
			expectedVersions = []string{"1.2.3", "2.0.0"}
			filteredReleases = allReleases
		})

		It("deduplicates by product version and returns one entry per version", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal("1.2.3"))
		})
	})

	Context("when the release type is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = string(releaseTypes[1])

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		It("returns the most recent version with that release type", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(expectedVersions[1])) // 2.3.4
		})

		Context("when the release type is invalid", func() {
			BeforeEach(func() {
				checkRequest.Source.ReleaseType = "not a valid release type"
			})

			It("returns an error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*release type.*one of"))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[0])))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[1])))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[2])))
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
			checkRequest.Source.ReleaseType = string(releaseTypes[1])

			filteredReleases = []pivnet.Release{allReleases[1]}
		})

		BeforeEach(func() {
			checkRequest.Source.ProductVersion = "C"
		})

		It("returns the newest release with that version without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(expectedVersions[1])) // 2.3.4
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

			semverOrderedReleases = []pivnet.Release{
				allReleases[1], // 2.3.4
				allReleases[2], // 1.2.4
				allReleases[0], // 1.2.3
			}

			checkRequest.Version = concourse.Version{
				ProductVersion: "1.2.3",
			}

			fakeSorter.SortBySemverReturns(semverOrderedReleases, nil)
		})

		It("returns in ascending semver order", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			// vs = 2.3.4, 1.2.4, 1.2.3; Since(1.2.3) then Reverse => 1.2.3, 1.2.4, 2.3.4
			Expect(response).To(HaveLen(3))
			Expect(response[0].ProductVersion).To(Equal("1.2.3"))
			Expect(response[1].ProductVersion).To(Equal("1.2.4"))
			Expect(response[2].ProductVersion).To(Equal("2.3.4"))

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

	Context("when sorting by last_updated", func() {
		var (
			releases []pivnet.Release
		)

		BeforeEach(func() {
			checkRequest.Source.SortBy = concourse.SortByLastUpdated

			releases = []pivnet.Release{
				allReleases[1], // 2.3.4
				allReleases[2], // 1.2.4
				allReleases[0], // 1.2.3
			}

			checkRequest.Version = concourse.Version{
				ProductVersion: "1.2.3",
			}

			fakeSorter.SortByLastUpdatedReturns(releases, nil)
		})

		It("returns invokes sort by update_at on sorter", func() {
			Expect(fakeSorter.SortByLastUpdatedCallCount()).To(Equal(0))

			_, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSorter.SortByLastUpdatedCallCount()).To(Equal(1))
		})

		Context("when sorting by semver returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakeSorter.SortByLastUpdatedReturns(nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})
	})
})
