package in_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/in"
	"github.com/pivotal-cf/pivnet-resource/in/infakes"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/versions"
)

var _ = Describe("In", func() {
	const (
		eulaSlug = "some-eula"
	)

	var (
		fakeFilter       *infakes.FakeFilter
		fakeDownloader   *infakes.FakeDownloader
		fakePivnetClient *infakes.FakePivnetClient
		fakeFileSummer   *infakes.FakeFileSummer
		fakeFileWriter   *infakes.FakeFileWriter

		productFiles         []pivnet.ProductFile
		filteredProductFiles []pivnet.ProductFile
		productFile1         pivnet.ProductFile
		productFile2         pivnet.ProductFile

		releaseDependencies []pivnet.ReleaseDependency
		releaseUpgradePaths []pivnet.ReleaseUpgradePath

		version                string
		fingerprint            string
		actualFingerprint      string
		versionWithFingerprint string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		release           pivnet.Release
		downloadFilepaths []string
		fileContentsMD5s  []string

		getReleaseErr          error
		actualFingerprintErr   error
		acceptEULAErr          error
		productFilesErr        error
		productFileErr         error
		downloadErr            error
		filterErr              error
		md5sumErr              error
		releaseDependenciesErr error
		releaseUpgradePathsErr error
	)

	BeforeEach(func() {
		fakeFilter = &infakes.FakeFilter{}
		fakeDownloader = &infakes.FakeDownloader{}
		fakePivnetClient = &infakes.FakePivnetClient{}
		fakeFileSummer = &infakes.FakeFileSummer{}
		fakeFileWriter = &infakes.FakeFileWriter{}

		getReleaseErr = nil
		actualFingerprintErr = nil
		acceptEULAErr = nil
		productFilesErr = nil
		productFileErr = nil
		filterErr = nil
		downloadErr = nil
		md5sumErr = nil
		releaseDependenciesErr = nil
		releaseUpgradePathsErr = nil

		version = "C"
		fingerprint = "fingerprint-0"
		actualFingerprint = fingerprint

		fileContentsMD5s = []string{
			"some-md5 1234",
			"some-md5 3456",
		}

		var err error
		versionWithFingerprint, err = versions.CombineVersionAndFingerprint(version, fingerprint)
		Expect(err).NotTo(HaveOccurred())

		downloadFilepaths = []string{
			"file-1234",
			"file-3456",
		}

		// The endpoint for all product files returns less metadata than the
		// individual product files, so we split them apart to differentiate them
		productFiles = []pivnet.ProductFile{
			{
				ID:           1234,
				Name:         "product file 1234",
				Description:  "some product file 1234",
				AWSObjectKey: downloadFilepaths[0],
			},
			{
				ID:           3456,
				Name:         "product file 3456",
				Description:  "some product file 3456",
				AWSObjectKey: downloadFilepaths[1],
			},
		}

		productFile1 = pivnet.ProductFile{
			ID:           productFiles[0].ID,
			Name:         productFiles[0].Name,
			Description:  productFiles[0].Description,
			AWSObjectKey: productFiles[0].AWSObjectKey,
			FileType:     pivnet.FileTypeSoftware,
			FileVersion:  "some-file-version 1234",
			MD5:          fileContentsMD5s[0],
			Links: &pivnet.Links{
				Download: map[string]string{
					"href": "foo",
				},
			},
		}

		productFile2 = pivnet.ProductFile{
			ID:           productFiles[1].ID,
			Name:         productFiles[1].Name,
			Description:  productFiles[1].Description,
			AWSObjectKey: productFiles[1].AWSObjectKey,
			FileType:     pivnet.FileTypeSoftware,
			FileVersion:  "some-file-version 3456",
			MD5:          fileContentsMD5s[1],
			Links: &pivnet.Links{
				Download: map[string]string{
					"href": "bar",
				},
			},
		}

		filteredProductFiles = []pivnet.ProductFile{productFile1, productFile2}

		file1URL := "some-file-path"

		release = pivnet.Release{
			Version:   version,
			UpdatedAt: actualFingerprint,
			ID:        1234,
			Links: &pivnet.Links{
				ProductFiles: map[string]string{
					"href": file1URL,
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
						Name: "some product",
					},
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
		release.UpdatedAt = actualFingerprint

		fakePivnetClient.GetReleaseReturns(release, getReleaseErr)
		fakePivnetClient.AcceptEULAReturns(acceptEULAErr)
		fakePivnetClient.ProductFilesForReleaseReturns(productFiles, productFilesErr)

		fakePivnetClient.ReleaseDependenciesReturns(releaseDependencies, releaseDependenciesErr)
		fakePivnetClient.ReleaseUpgradePathsReturns(releaseUpgradePaths, releaseUpgradePathsErr)

		fakePivnetClient.ProductFileForReleaseStub = func(
			productSlug string,
			releaseID int,
			productFileID int,
		) (pivnet.ProductFile, error) {
			if productFileErr != nil {
				return pivnet.ProductFile{}, productFileErr
			}

			switch productFileID {
			case productFile1.ID:
				return productFile1, nil
			case productFile2.ID:
				return productFile2, nil
			}

			Fail(fmt.Sprintf("unexpected productFileID: %d", productFileID))
			return pivnet.ProductFile{}, nil
		}

		fakeFilter.ProductFileNamesByGlobsReturns(filteredProductFiles, filterErr)
		fakeDownloader.DownloadReturns(downloadFilepaths, downloadErr)
		fakeFileSummer.SumFileStub = func(path string) (string, error) {
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

		logging := log.New(ioutil.Discard, "doesn't matter", 0)

		inCommand = in.NewInCommand(
			logging,
			fakePivnetClient,
			fakeFilter,
			fakeDownloader,
			fakeFileSummer,
			fakeFileWriter,
		)
	})

	It("invokes the version file writer with downloaded version and fingerprint", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteVersionFileCallCount()).To(Equal(1))
		Expect(fakeFileWriter.WriteVersionFileArgsForCall(0)).To(Equal(versionWithFingerprint))
	})

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
			Expect(writtenMetadata.ProductFiles[i].MD5).To(Equal(p.MD5))
			Expect(writtenMetadata.ProductFiles[i].UploadAs).To(BeEmpty())
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
			Expect(writtenMetadata.Dependencies[i].Release.Product.Name).To(Equal(d.Release.Product.Name))
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

	It("invokes the json metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataJSONFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataJSONFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.Version).To(Equal(version))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		pFiles := []pivnet.ProductFile{
			productFile1,
			productFile2,
		}

		validateProductFilesMetadata(invokedMetadata, pFiles)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		validateReleaseUpgradePathsMetadata(invokedMetadata, releaseUpgradePaths)
	})

	It("invokes the yaml metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataYAMLFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataYAMLFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.Version).To(Equal(version))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		pFiles := []pivnet.ProductFile{
			productFile1,
			productFile2,
		}

		validateProductFilesMetadata(invokedMetadata, pFiles)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		validateReleaseUpgradePathsMetadata(invokedMetadata, releaseUpgradePaths)
	})

	It("downloads all files (nil globs acts like *)", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFilter.ProductFileNamesByGlobsCallCount()).To(Equal(1))
		Expect(fakePivnetClient.ProductFileForReleaseCallCount()).To(Equal(len(productFiles)))
		Expect(fakeFileSummer.SumFileCallCount()).To(Equal(len(downloadFilepaths)))
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

	Context("when getting individual product file returns error", func() {
		BeforeEach(func() {
			productFileErr = fmt.Errorf("some product file error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(productFileErr))
		})
	})

	Context("when globs are provided", func() {
		BeforeEach(func() {
			inRequest.Params.Globs = []string{"some*glob", "other*glob"}
		})

		It("downloads files", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFilter.ProductFileNamesByGlobsCallCount()).To(Equal(1))
			Expect(fakePivnetClient.ProductFileForReleaseCallCount()).To(Equal(len(productFiles)))
			Expect(fakeFileSummer.SumFileCallCount()).To(Equal(len(downloadFilepaths)))
		})

		It("includes md5 when invoking metadata writer", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileWriter.WriteMetadataYAMLFileCallCount()).To(Equal(1))
			invokedMetadata := fakeFileWriter.WriteMetadataYAMLFileArgsForCall(0)

			Expect(invokedMetadata.Release).NotTo(BeNil())
			Expect(invokedMetadata.Release.Version).To(Equal(version))
			Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

			pFiles := []pivnet.ProductFile{
				productFile1,
				productFile2,
			}

			validateProductFilesMetadata(invokedMetadata, pFiles)
			validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		})

		Context("when the file type is not 'Software'", func() {
			BeforeEach(func() {
				productFile2.FileType = "not software"
				fileContentsMD5s[1] = "not interested"
			})

			It("ignores MD5", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when getting a product file returns error", func() {
			BeforeEach(func() {
				productFileErr = fmt.Errorf("some product file error")
			})

			It("returns error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(productFileErr))
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

			Context("when the error is no globs match", func() {
				BeforeEach(func() {
					filterErr = filter.ErrNoMatch{}
				})

				It("returns the error", func() {
					_, err := inCommand.Run(inRequest)
					Expect(err).To(HaveOccurred())
				})

				Context("when globs are nil", func() {
					BeforeEach(func() {
						inRequest.Params.Globs = nil
					})

					It("does not return an error", func() {
						_, err := inCommand.Run(inRequest)
						Expect(err).NotTo(HaveOccurred())
					})
				})
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
})
