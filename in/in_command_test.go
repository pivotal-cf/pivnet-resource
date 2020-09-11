package in_test

import (
	"fmt"
	"log"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-resource/v2/concourse"
	"github.com/pivotal-cf/pivnet-resource/v2/in"
	"github.com/pivotal-cf/pivnet-resource/v2/in/infakes"
	"github.com/pivotal-cf/pivnet-resource/v2/metadata"
	"github.com/pivotal-cf/pivnet-resource/v2/versions"
)

var _ = Describe("In", func() {
	const (
		eulaSlug = "some-eula"
	)

	var (
		fakeLogger logger.Logger

		fakeFilter           *infakes.FakeFilter
		fakeDownloader       *infakes.FakeDownloader
		fakePivnetClient     *infakes.FakePivnetClient
		fakeSHA256FileSummer *infakes.FakeFileSummer
		fakeMD5FileSummer    *infakes.FakeFileSummer
		fakeFileWriter       *infakes.FakeFileWriter
		fakeArchive          *infakes.FakeArchive

		fileGroups []pivnet.FileGroup

		releaseProductFiles    []pivnet.ProductFile
		fileGroup1ProductFiles []pivnet.ProductFile
		fileGroup2ProductFiles []pivnet.ProductFile

		filteredProductFiles []pivnet.ProductFile

		artifactReferences  []pivnet.ArtifactReference

		releaseDependencies   []pivnet.ReleaseDependency
		dependencySpecifiers  []pivnet.DependencySpecifier
		releaseUpgradePaths   []pivnet.ReleaseUpgradePath
		upgradePathSpecifiers []pivnet.UpgradePathSpecifier

		version                string
		fingerprint            string
		actualFingerprint      string
		versionWithFingerprint string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		release             pivnet.Release
		downloadFilepaths   []string
		fileContentsSHA256s []string
		fileContentsMD5s    []string

		getReleaseErr            error
		acceptEULAErr            error
		productFilesErr          error
		downloadErr              error
		filterErr                error
		sha256sumErr             error
		md5sumErr                error
		releaseDependenciesErr   error
		dependencySpecifiersErr  error
		releaseUpgradePathsErr   error
		upgradePathSpecifiersErr error
		fileGroupsErr            error
		artifactReferencesErr    error
	)

	BeforeEach(func() {
		fakeFilter = &infakes.FakeFilter{}
		fakeDownloader = &infakes.FakeDownloader{}
		fakePivnetClient = &infakes.FakePivnetClient{}
		fakeSHA256FileSummer = &infakes.FakeFileSummer{}
		fakeMD5FileSummer = &infakes.FakeFileSummer{}
		fakeFileWriter = &infakes.FakeFileWriter{}
		fakeArchive = &infakes.FakeArchive{}

		getReleaseErr = nil
		acceptEULAErr = nil
		productFilesErr = nil
		filterErr = nil
		downloadErr = nil
		sha256sumErr = nil
		md5sumErr = nil
		releaseDependenciesErr = nil
		dependencySpecifiersErr = nil
		releaseUpgradePathsErr = nil
		upgradePathSpecifiersErr = nil
		fileGroupsErr = nil
		artifactReferencesErr = nil

		version = "C"
		fingerprint = "fingerprint-0"
		actualFingerprint = fingerprint

		fileContentsSHA256s = []string{
			"some-sha256 1234",
			"some-sha256 3456",
			"some-sha256 4567",
			"some-sha256 5678",
		}

		fileContentsMD5s = []string{
			"some-md5 1234",
			"some-md5 3456",
			"some-md5 4567",
			"some-md5 5678",
		}

		var err error
		versionWithFingerprint, err = versions.CombineVersionAndFingerprint(version, fingerprint)
		Expect(err).NotTo(HaveOccurred())

		downloadFilepaths = []string{
			"file-1234",
			"file-3456",
			"file-4567",
			"file-5678",
		}

		// The endpoint for all product files returns less metadata than the
		// individual product files, so we split them apart to differentiate them
		releaseProductFiles = []pivnet.ProductFile{
			{
				ID:           1234,
				Name:         "product file 1234",
				AWSObjectKey: downloadFilepaths[0],
				FileType:     pivnet.FileTypeSoftware,
				FileVersion:  "some-file-version 1234",
				SHA256:       fileContentsSHA256s[0],
				MD5:          fileContentsMD5s[0],
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "foo",
					},
				},
			},
			{
				ID:           3456,
				Name:         "product file 3456",
				AWSObjectKey: downloadFilepaths[1],
				FileType:     pivnet.FileTypeSoftware,
				FileVersion:  "some-file-version 3456",
				SHA256:       fileContentsSHA256s[1],
				MD5:          fileContentsMD5s[1],
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "bar",
					},
				},
			},
		}

		fileGroup1ProductFiles = []pivnet.ProductFile{
			{
				ID:           4567,
				Name:         "product file 4567",
				AWSObjectKey: downloadFilepaths[2],
				FileType:     pivnet.FileTypeSoftware,
				FileVersion:  "some-file-version 4567",
				SHA256:       fileContentsSHA256s[2],
				MD5:          fileContentsMD5s[2],
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "bar",
					},
				},
			},
		}

		fileGroup2ProductFiles = []pivnet.ProductFile{
			{
				ID:           5678,
				Name:         "product file 5678",
				AWSObjectKey: downloadFilepaths[3],
				FileType:     pivnet.FileTypeSoftware,
				FileVersion:  "some-file-version 5678",
				SHA256:       fileContentsSHA256s[3],
				MD5:          fileContentsMD5s[3],
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "bar",
					},
				},
			},
		}

		filteredProductFiles = []pivnet.ProductFile{
			releaseProductFiles[0],
			releaseProductFiles[1],
			fileGroup1ProductFiles[0],
			fileGroup2ProductFiles[0],
		}

		fileGroups = []pivnet.FileGroup{
			{
				ID:   4321,
				Name: "fg1",
				ProductFiles: []pivnet.ProductFile{
					fileGroup1ProductFiles[0],
				},
			},
			{
				ID:   5432,
				Name: "fg2",
				ProductFiles: []pivnet.ProductFile{
					fileGroup2ProductFiles[0],
				},
			},
		}

		artifactReferences = []pivnet.ArtifactReference{
			{
				ID:                 101,
				Name:               "artifact1",
				ArtifactPath:       "my/path:1",
				Digest:             "mydigest1",
				Description:        "my description 1",
				DocsURL:            "my.docs.url:1",
				SystemRequirements: []string{"a", "b", "c"},
			},
			{
				ID:                 102,
				Name:               "artifact2",
				ArtifactPath:       "my/path:2",
				Digest:             "mydigest2",
				Description:        "my description 2",
				DocsURL:            "my.docs.url:2",
				SystemRequirements: []string{"d", "e", "f"},
			},
		}

		release = pivnet.Release{
			Version:                version,
			SoftwareFilesUpdatedAt: actualFingerprint,
			ID:                     1234,
			Links: &pivnet.Links{
				ProductFiles: map[string]string{
					"href": "some-file-path",
				},
			},
			EULA: &pivnet.EULA{
				Slug: eulaSlug,
				ID:   1234,
				Name: "some EULA",
			},
		}

		releaseDependencies = []pivnet.ReleaseDependency{
			{
				Release: pivnet.DependentRelease{
					ID:      56,
					Version: "dependent release 56",
					Product: pivnet.Product{
						ID:   67,
						Slug: "some-slug",
						Name: "some product",
					},
				},
			},
		}

		dependencySpecifiers = []pivnet.DependencySpecifier{
			{
				ID:        56,
				Specifier: "1.2.*",
				Product: pivnet.Product{
					Slug: "some-product",
				},
			},
		}

		releaseUpgradePaths = []pivnet.ReleaseUpgradePath{
			{
				Release: pivnet.UpgradePathRelease{
					ID:      56,
					Version: "upgrade release 56",
				},
			},
		}

		upgradePathSpecifiers = []pivnet.UpgradePathSpecifier{
			{
				ID:        56,
				Specifier: "1.2.*",
			},
		}

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
			},
			Version: concourse.Version{
				ProductVersion: versionWithFingerprint,
			},
		}
	})

	JustBeforeEach(func() {
		release.SoftwareFilesUpdatedAt = actualFingerprint

		fakePivnetClient.GetReleaseReturns(release, getReleaseErr)
		fakePivnetClient.AcceptEULAReturns(acceptEULAErr)
		fakePivnetClient.ProductFilesForReleaseReturns(releaseProductFiles, productFilesErr)

		fakePivnetClient.ReleaseDependenciesReturns(releaseDependencies, releaseDependenciesErr)
		fakePivnetClient.DependencySpecifiersReturns(dependencySpecifiers, dependencySpecifiersErr)
		fakePivnetClient.ReleaseUpgradePathsReturns(releaseUpgradePaths, releaseUpgradePathsErr)
		fakePivnetClient.UpgradePathSpecifiersReturns(upgradePathSpecifiers, upgradePathSpecifiersErr)
		fakePivnetClient.FileGroupsForReleaseReturns(fileGroups, fileGroupsErr)
		fakePivnetClient.ArtifactReferencesForReleaseReturns(artifactReferences, artifactReferencesErr)

		fakeFilter.ProductFileKeysByGlobsReturns(filteredProductFiles, filterErr)
		fakeDownloader.DownloadReturns(downloadFilepaths, downloadErr)
		fakeSHA256FileSummer.SumFileStub = func(path string) (string, error) {
			if sha256sumErr != nil {
				return "", sha256sumErr
			}

			for i, f := range downloadFilepaths {
				if strings.HasSuffix(path, f) {
					return fileContentsSHA256s[i], nil
				}
			}

			Fail(fmt.Sprintf("unexpected path: %s", path))
			return "", nil
		}
		fakeMD5FileSummer.SumFileStub = func(path string) (string, error) {
			if md5sumErr != nil {
				return "", md5sumErr
			}

			for i, f := range downloadFilepaths {
				if strings.HasSuffix(path, f) {
					return fileContentsMD5s[i], nil
				}
			}

			Fail(fmt.Sprintf("unexpected path: %s", path))
			return "", nil
		}

		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		inCommand = in.NewInCommand(
			fakeLogger,
			fakePivnetClient,
			fakeFilter,
			fakeDownloader,
			fakeSHA256FileSummer,
			fakeMD5FileSummer,
			fakeFileWriter,
			fakeArchive,
		)
	})

	It("invokes the version file writer with downloaded version and fingerprint", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteVersionFileCallCount()).To(Equal(1))
		Expect(fakeFileWriter.WriteVersionFileArgsForCall(0)).To(Equal(versionWithFingerprint))
	})

	It("invokes the json metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataJSONFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataJSONFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.ID).To(Equal(release.ID))
		Expect(invokedMetadata.Release.Version).To(Equal(version))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		validateReleaseProductFilesMetadata(invokedMetadata, releaseProductFiles)
		validateProductFilesMetadata(invokedMetadata, filteredProductFiles)
		validateFileGroupsMetadata(invokedMetadata, fileGroups)
		validateArtifactReferencesMetadata(invokedMetadata, artifactReferences)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		validateDependencySpecifiersMetadata(invokedMetadata, dependencySpecifiers)
		validateReleaseUpgradePathsMetadata(invokedMetadata, releaseUpgradePaths)
		validateUpgradePathSpecifiersMetadata(invokedMetadata, upgradePathSpecifiers)
	})

	It("invokes the yaml metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataYAMLFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataYAMLFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.ID).To(Equal(release.ID))
		Expect(invokedMetadata.Release.Version).To(Equal(version))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		validateReleaseProductFilesMetadata(invokedMetadata, releaseProductFiles)
		validateProductFilesMetadata(invokedMetadata, filteredProductFiles)
		validateFileGroupsMetadata(invokedMetadata, fileGroups)
		validateArtifactReferencesMetadata(invokedMetadata, artifactReferences)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		validateDependencySpecifiersMetadata(invokedMetadata, dependencySpecifiers)
		validateReleaseUpgradePathsMetadata(invokedMetadata, releaseUpgradePaths)
		validateUpgradePathSpecifiersMetadata(invokedMetadata, upgradePathSpecifiers)
	})

	It("downloads all files (nil globs acts like *)", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakePivnetClient.ProductFilesForReleaseCallCount()).To(Equal(1))
		Expect(fakePivnetClient.FileGroupsForReleaseCallCount()).To(Equal(1))

		Expect(fakeFilter.ProductFileKeysByGlobsCallCount()).To(Equal(0))

		expectedProductFiles := releaseProductFiles
		expectedProductFiles = append(expectedProductFiles, fileGroup1ProductFiles[0])
		expectedProductFiles = append(expectedProductFiles, fileGroup2ProductFiles[0])

		Expect(fakeDownloader.DownloadCallCount()).To(Equal(1))
		invokedProductFiles, _, _ := fakeDownloader.DownloadArgsForCall(0)
		Expect(invokedProductFiles).To(Equal(filteredProductFiles))

		Expect(fakeSHA256FileSummer.SumFileCallCount() + fakeMD5FileSummer.SumFileCallCount()).To(Equal(len(downloadFilepaths)))
	})

	Context("when version is provided without fingerprint", func() {
		BeforeEach(func() {
			inRequest.Version = concourse.Version{
				ProductVersion: version,
			}
		})

		It("returns without error (does not compare against actual fingerprint)", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when getting release returns error", func() {
		BeforeEach(func() {
			getReleaseErr = fmt.Errorf("some release error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getReleaseErr))
		})
	})

	Context("when actual fingerprint is different than provided", func() {
		BeforeEach(func() {
			actualFingerprint = "different fingerprint"
		})

		It("returns the error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(
				".*provided.*'%s'.*actual.*'%s'.*",
				fingerprint,
				actualFingerprint,
			))
		})
	})

	Context("when accepting EULA returns error", func() {
		BeforeEach(func() {
			acceptEULAErr = fmt.Errorf("some eula error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(acceptEULAErr))
		})
	})

	Context("when getting file groups returns error", func() {
		BeforeEach(func() {
			fileGroupsErr = fmt.Errorf("some file group error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(fileGroupsErr))
		})
	})

	Context("when getting artifact references returns error", func() {
		BeforeEach(func() {
			artifactReferencesErr = fmt.Errorf("some artifact reference error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(artifactReferencesErr))
		})
	})

	Context("when getting product files returns error", func() {
		BeforeEach(func() {
			productFilesErr = fmt.Errorf("some product files error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(productFilesErr))
		})
	})

	Describe("when globs are provided", func() {
		BeforeEach(func() {
			inRequest.Params.Globs = []string{"some*glob", "other*glob"}
		})

		It("downloads files, filtering by globs", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFilter.ProductFileKeysByGlobsCallCount()).To(Equal(1))
			Expect(fakeSHA256FileSummer.SumFileCallCount() + fakeMD5FileSummer.SumFileCallCount()).To(Equal(len(downloadFilepaths)))
		})

		Context("when the file type is not 'Software'", func() {
			BeforeEach(func() {
				releaseProductFiles[1].FileType = "not software"
				fileContentsSHA256s[1] = "this would fail if type was software"
				fileContentsMD5s[1] = "this would fail if type was software"
			})

			It("ignores SHA256", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores MD5", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when filtering an returns error", func() {
			BeforeEach(func() {
				filterErr = fmt.Errorf("some filter error")
			})

			It("returns the error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(filterErr))
			})
		})

		Context("when downloading files returns an error", func() {
			BeforeEach(func() {
				downloadErr = fmt.Errorf("some download error")
			})

			It("returns the error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(downloadErr))
			})
		})

		Context("When SHA256 is supplied", func() {
			BeforeEach(func() {
				md5sumErr = fmt.Errorf("some md5 err error")
			})

			It("ignores MD5", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when calculating sha256 sum of file returns an error", func() {
				BeforeEach(func() {
					sha256sumErr = fmt.Errorf("some sha256 err error")
				})

				It("returns the error", func() {
					_, err := inCommand.Run(inRequest)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(sha256sumErr))
				})
			})

			Context("when the SHA256 sum does not match", func() {
				BeforeEach(func() {
					fileContentsSHA256s[0] = "incorrect sha256"
				})

				It("returns an error", func() {
					_, err := inCommand.Run(inRequest)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("When SHA256 is not supplied", func() {
			BeforeEach(func() {
				releaseProductFiles[0].SHA256 = ""
				fileContentsSHA256s[0] = ""
			})

			It("does not return an error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when calculating md5 sum of file returns an error", func() {
				BeforeEach(func() {
					md5sumErr = fmt.Errorf("some md5 err error")
				})

				It("returns the error", func() {
					_, err := inCommand.Run(inRequest)
					Expect(err).To(HaveOccurred())

					Expect(err).To(Equal(md5sumErr))
				})
			})

			Context("when the MD5 does not match", func() {
				BeforeEach(func() {
					fileContentsMD5s[0] = "incorrect md5"
				})

				It("returns an error", func() {
					_, err := inCommand.Run(inRequest)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("when unpack is set", func() {
		BeforeEach(func() {
			inRequest.Params.Unpack = true

		})

		It("downloads files and extracts archive", func() {
			fakeArchive.MimetypeReturns("application/gzip")
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())
		})

		It("downloads files and continues when file is not an archive", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when getting release dependencies returns an error", func() {
		BeforeEach(func() {
			releaseDependenciesErr = fmt.Errorf("some release dependencies error")
		})

		It("returns the error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(releaseDependenciesErr))
		})
	})

	Context("when getting dependency specifiers returns an error", func() {
		BeforeEach(func() {
			dependencySpecifiersErr = fmt.Errorf("some dependency specifiers error")
		})

		It("returns the error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(dependencySpecifiersErr))
		})
	})

	Context("when getting release upgrade paths returns an error", func() {
		BeforeEach(func() {
			releaseUpgradePathsErr = fmt.Errorf("some release upgrade paths error")
		})

		It("returns the error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(releaseUpgradePathsErr))
		})
	})

	Context("when getting upgrade path specifiers returns an error", func() {
		BeforeEach(func() {
			upgradePathSpecifiersErr = fmt.Errorf("some upgrade path specifiers error")
		})

		It("returns the error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(upgradePathSpecifiersErr))
		})
	})
})

var validateFileGroupsMetadata = func(
	writtenMetadata metadata.Metadata,
	fileGroups []pivnet.FileGroup,
) {
	Expect(writtenMetadata.FileGroups).To(HaveLen(len(fileGroups)))

	for i, fg := range fileGroups {
		Expect(writtenMetadata.FileGroups[i].ID).To(Equal(fg.ID))
		Expect(writtenMetadata.FileGroups[i].Name).To(Equal(fg.Name))

		for j, p := range fg.ProductFiles {
			Expect(writtenMetadata.FileGroups[i].ProductFiles[j].ID).To(Equal(p.ID))
		}
	}
}

var validateReleaseProductFilesMetadata = func(
	writtenMetadata metadata.Metadata,
	pF []pivnet.ProductFile,
) {
	Expect(writtenMetadata.Release.ProductFiles).To(HaveLen(len(pF)))

	for i, p := range pF {
		Expect(writtenMetadata.Release.ProductFiles[i].ID).To(Equal(p.ID))
	}
}

var validateProductFilesMetadata = func(
	writtenMetadata metadata.Metadata,
	pF []pivnet.ProductFile,
) {
	Expect(writtenMetadata.ProductFiles).To(HaveLen(len(pF)))

	for i, p := range pF {
		Expect(writtenMetadata.ProductFiles[i].File).To(Equal(p.Name))
		Expect(writtenMetadata.ProductFiles[i].Description).To(Equal(p.Description))
		Expect(writtenMetadata.ProductFiles[i].ID).To(Equal(p.ID))
		Expect(writtenMetadata.ProductFiles[i].AWSObjectKey).To(Equal(p.AWSObjectKey))
		Expect(writtenMetadata.ProductFiles[i].FileType).To(Equal(p.FileType))
		Expect(writtenMetadata.ProductFiles[i].FileVersion).To(Equal(p.FileVersion))
		Expect(writtenMetadata.ProductFiles[i].SHA256).To(Equal(p.SHA256))
		Expect(writtenMetadata.ProductFiles[i].MD5).To(Equal(p.MD5))
		Expect(writtenMetadata.ProductFiles[i].UploadAs).To(BeEmpty())
		Expect(writtenMetadata.ProductFiles[i].DocsURL).To(Equal(p.DocsURL))
		Expect(writtenMetadata.ProductFiles[i].SystemRequirements).To(Equal(p.SystemRequirements))
	}
}

var validateArtifactReferencesMetadata = func(
	writtenMetadata metadata.Metadata,
	artifactReferences []pivnet.ArtifactReference,
) {
	Expect(writtenMetadata.ArtifactReferences).To(HaveLen(len(artifactReferences)))

	for i, artifactReference := range artifactReferences {
		Expect(writtenMetadata.ArtifactReferences[i].ID).To(Equal(artifactReference.ID))
		Expect(writtenMetadata.ArtifactReferences[i].Name).To(Equal(artifactReference.Name))
		Expect(writtenMetadata.ArtifactReferences[i].ArtifactPath).To(Equal(artifactReference.ArtifactPath))
		Expect(writtenMetadata.ArtifactReferences[i].Digest).To(Equal(artifactReference.Digest))
		Expect(writtenMetadata.ArtifactReferences[i].Description).To(Equal(artifactReference.Description))
		Expect(writtenMetadata.ArtifactReferences[i].DocsURL).To(Equal(artifactReference.DocsURL))
		Expect(writtenMetadata.ArtifactReferences[i].SystemRequirements).To(Equal(artifactReference.SystemRequirements))
	}
}

var validateReleaseDependenciesMetadata = func(
	writtenMetadata metadata.Metadata,
	dependencies []pivnet.ReleaseDependency,
) {
	Expect(writtenMetadata.Dependencies).To(HaveLen(len(dependencies)))

	for i, d := range dependencies {
		Expect(writtenMetadata.Dependencies[i].Release.ID).To(Equal(d.Release.ID))
		Expect(writtenMetadata.Dependencies[i].Release.Version).To(Equal(d.Release.Version))
		Expect(writtenMetadata.Dependencies[i].Release.Product.ID).To(Equal(d.Release.Product.ID))
		Expect(writtenMetadata.Dependencies[i].Release.Product.Slug).To(Equal(d.Release.Product.Slug))
		Expect(writtenMetadata.Dependencies[i].Release.Product.Name).To(Equal(d.Release.Product.Name))
	}
}

var validateDependencySpecifiersMetadata = func(
	writtenMetadata metadata.Metadata,
	dependencySpecifiers []pivnet.DependencySpecifier,
) {
	Expect(writtenMetadata.DependencySpecifiers).To(HaveLen(len(dependencySpecifiers)))

	for i, d := range dependencySpecifiers {
		Expect(writtenMetadata.DependencySpecifiers[i].ID).To(Equal(d.ID))
		Expect(writtenMetadata.DependencySpecifiers[i].Specifier).To(Equal(d.Specifier))
		Expect(writtenMetadata.DependencySpecifiers[i].ProductSlug).To(Equal(d.Product.Slug))
	}
}

var validateReleaseUpgradePathsMetadata = func(
	writtenMetadata metadata.Metadata,
	upgradePaths []pivnet.ReleaseUpgradePath,
) {
	Expect(writtenMetadata.UpgradePaths).To(HaveLen(len(upgradePaths)))

	for i, d := range upgradePaths {
		Expect(writtenMetadata.UpgradePaths[i].ID).To(Equal(d.Release.ID))
		Expect(writtenMetadata.UpgradePaths[i].Version).To(Equal(d.Release.Version))
	}
}

var validateUpgradePathSpecifiersMetadata = func(
	writtenMetadata metadata.Metadata,
	upgradePathSpecifiers []pivnet.UpgradePathSpecifier,
) {
	Expect(writtenMetadata.UpgradePathSpecifiers).To(HaveLen(len(upgradePathSpecifiers)))

	for i, d := range upgradePathSpecifiers {
		Expect(writtenMetadata.UpgradePathSpecifiers[i].ID).To(Equal(d.ID))
		Expect(writtenMetadata.UpgradePathSpecifiers[i].Specifier).To(Equal(d.Specifier))
	}
}
