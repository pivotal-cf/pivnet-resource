package in_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader/downloaderfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter/filterfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/gp/gpfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/in/filesystem/filesystemfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5sum/md5sumfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("In", func() {
	const (
		eulaSlug = "some-eula"
	)

	var (
		fakeFilter       *filterfakes.FakeFilter
		fakeDownloader   *downloaderfakes.FakeDownloader
		fakePivnetClient *gpfakes.FakeClient
		fakeFileSummer   *md5sumfakes.FakeFileSummer
		fakeFileWriter   *filesystemfakes.FakeFileWriter

		productFiles []pivnet.ProductFile
		productFile1 pivnet.ProductFile
		productFile2 pivnet.ProductFile

		releaseDependencies []pivnet.ReleaseDependency

		productVersion  string
		etag            string
		versionWithETag string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		release           pivnet.Release
		downloadFilepaths []string
		fileContentsMD5s  []string

		getReleaseErr          error
		acceptEULAErr          error
		getProductFilesErr     error
		getProductFileErr      error
		downloadErr            error
		downloadLinksByGlobErr error
		md5sumErr              error
		releaseDependenciesErr error
	)

	BeforeEach(func() {
		fakeFilter = &filterfakes.FakeFilter{}
		fakeDownloader = &downloaderfakes.FakeDownloader{}
		fakePivnetClient = &gpfakes.FakeClient{}
		fakeFileSummer = &md5sumfakes.FakeFileSummer{}
		fakeFileWriter = &filesystemfakes.FakeFileWriter{}

		getReleaseErr = nil
		acceptEULAErr = nil
		getProductFilesErr = nil
		getProductFileErr = nil
		downloadLinksByGlobErr = nil
		downloadErr = nil
		md5sumErr = nil
		releaseDependenciesErr = nil

		productVersion = "C"
		etag = "etag-0"

		fileContentsMD5s = []string{
			"some-md5 1234",
			"some-md5 3456",
		}

		var err error
		versionWithETag, err = versions.CombineVersionAndETag(productVersion, etag)
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
			FileVersion:  "some-file-version 1234",
			FileType:     "some-file-type 1234",
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
			FileType:     "some-file-type 1234",
			FileVersion:  "some-file-version 3456",
			MD5:          fileContentsMD5s[1],
			Links: &pivnet.Links{
				Download: map[string]string{
					"href": "bar",
				},
			},
		}

		file1URL := "some-file-path"

		release = pivnet.Release{
			Version: productVersion,
			ID:      1234,
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

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
			},
			Version: concourse.Version{
				ProductVersion: versionWithETag,
			},
		}
	})

	JustBeforeEach(func() {
		fakePivnetClient.GetReleaseReturns(release, getReleaseErr)
		fakePivnetClient.AcceptEULAReturns(acceptEULAErr)
		fakePivnetClient.GetProductFilesReturns(productFiles, getProductFilesErr)

		fakePivnetClient.ReleaseDependenciesReturns(releaseDependencies, releaseDependenciesErr)

		fakePivnetClient.GetProductFileStub = func(
			productSlug string,
			releaseID int,
			productFileID int,
		) (pivnet.ProductFile, error) {
			if getProductFileErr != nil {
				return pivnet.ProductFile{}, getProductFileErr
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

		fakeFilter.DownloadLinksByGlobReturns(map[string]string{}, downloadLinksByGlobErr)
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

		testLogger = lagertest.NewTestLogger("in unit tests")

		inCommand = in.NewInCommand(
			testLogger,
			fakePivnetClient,
			fakeFilter,
			fakeDownloader,
			fakeFileSummer,
			fakeFileWriter,
		)
	})

	It("invokes the version file writer with downloaded version and etag", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteVersionFileCallCount()).To(Equal(1))
		Expect(fakeFileWriter.WriteVersionFileArgsForCall(0)).To(Equal(versionWithETag))
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

	It("invokes the json metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataJSONFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataJSONFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.Version).To(Equal(productVersion))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		pFiles := []pivnet.ProductFile{
			productFile1,
			productFile2,
		}

		validateProductFilesMetadata(invokedMetadata, pFiles)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
	})

	It("invokes the yaml metadata file writer with correct metadata", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFileWriter.WriteMetadataYAMLFileCallCount()).To(Equal(1))
		invokedMetadata := fakeFileWriter.WriteMetadataYAMLFileArgsForCall(0)

		Expect(invokedMetadata.Release).NotTo(BeNil())
		Expect(invokedMetadata.Release.Version).To(Equal(productVersion))
		Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

		pFiles := []pivnet.ProductFile{
			productFile1,
			productFile2,
		}

		validateProductFilesMetadata(invokedMetadata, pFiles)
		validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
	})

	It("does not download any of the files in the specified release", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeDownloader.DownloadCallCount()).To(Equal(0))
	})

	Context("when version is provided without etag", func() {
		BeforeEach(func() {
			inRequest.Version = concourse.Version{
				ProductVersion: productVersion,
			}
		})

		It("returns without error", func() {
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
			getProductFilesErr = fmt.Errorf("some product files error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getProductFilesErr))
		})
	})

	Context("when getting individual product file returns error", func() {
		BeforeEach(func() {
			getProductFileErr = fmt.Errorf("some product file error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getProductFileErr))
		})
	})

	Context("when globs are provided", func() {
		BeforeEach(func() {
			inRequest.Params.Globs = []string{"some*glob", "other*glob"}
		})

		It("downloads files", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFilter.DownloadLinksCallCount()).To(Equal(1))
			Expect(fakeFilter.DownloadLinksByGlobCallCount()).To(Equal(1))
			Expect(fakePivnetClient.GetProductFileCallCount()).To(Equal(len(productFiles)))
			Expect(fakeFileSummer.SumFileCallCount()).To(Equal(len(downloadFilepaths)))
		})

		It("includes md5 when invoking metadata writer", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileWriter.WriteMetadataYAMLFileCallCount()).To(Equal(1))
			invokedMetadata := fakeFileWriter.WriteMetadataYAMLFileArgsForCall(0)

			Expect(invokedMetadata.Release).NotTo(BeNil())
			Expect(invokedMetadata.Release.Version).To(Equal(productVersion))
			Expect(invokedMetadata.Release.EULASlug).To(Equal(eulaSlug))

			pFiles := []pivnet.ProductFile{
				productFile1,
				productFile2,
			}

			validateProductFilesMetadata(invokedMetadata, pFiles)
			validateReleaseDependenciesMetadata(invokedMetadata, releaseDependencies)
		})

		Context("when getting a product file returns error", func() {
			BeforeEach(func() {
				getProductFileErr = fmt.Errorf("some product file error")
			})

			It("returns error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(getProductFileErr))
			})
		})

		Context("when filtering download links returns error", func() {
			BeforeEach(func() {
				downloadLinksByGlobErr = fmt.Errorf("some filter error")
			})

			It("returns error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(downloadLinksByGlobErr))
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
})
