package in_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader/downloaderfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/filter/filterfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet/pivnetfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("In", func() {
	var (
		fakeFilter       *filterfakes.FakeFilter
		fakeDownloader   *downloaderfakes.FakeDownloader
		fakePivnetClient *pivnetfakes.FakeClient

		productFiles         []pivnet.ProductFile
		productFilesResponse pivnet.ProductFiles

		downloadDir string

		ginkgoLogger logger.Logger

		productVersion  string
		etag            string
		versionWithETag string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		release       pivnet.Release
		getReleaseErr error

		acceptEULAErr      error
		getProductFilesErr error
		getProductFileErr  error

		downloadLinksByGlobErr error
	)

	BeforeEach(func() {
		fakeFilter = &filterfakes.FakeFilter{}
		fakeDownloader = &downloaderfakes.FakeDownloader{}
		fakePivnetClient = &pivnetfakes.FakeClient{}

		getReleaseErr = nil
		acceptEULAErr = nil
		getProductFilesErr = nil
		getProductFileErr = nil
		downloadLinksByGlobErr = nil

		productVersion = "C"
		etag = "etag-0"

		var err error
		versionWithETag, err = versions.CombineVersionAndETag(productVersion, etag)
		Expect(err).NotTo(HaveOccurred())

		file1URL := "some-file-path"
		productFiles = []pivnet.ProductFile{
			{
				ID:           1234,
				Name:         "product file 1234",
				Description:  "some product file 1234",
				AWSObjectKey: "some-key 1234",
				FileType:     "some-file-type 1234",
				FileVersion:  "some-file-version 1234",
				MD5:          "some-md5 1234",
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "foo",
					},
				},
			},
			{
				ID:           3456,
				Name:         "product file 3456",
				Description:  "some product file 3456",
				AWSObjectKey: "some-key 3456",
				FileType:     "some-file-type 3456",
				FileVersion:  "some-file-version 3456",
				MD5:          "some-md5 3456",
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "bar",
					},
				},
			},
		}
		productFilesResponse = pivnet.ProductFiles{
			ProductFiles: productFiles,
		}

		release = pivnet.Release{
			Version: productVersion,
			ID:      1234,
			Links: &pivnet.Links{
				ProductFiles: map[string]string{
					"href": file1URL,
				},
			},
		}

		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
			},
			Version: concourse.Version{
				versionWithETag,
			},
		}
	})

	JustBeforeEach(func() {
		fakePivnetClient.GetReleaseReturns(release, getReleaseErr)
		fakePivnetClient.AcceptEULAReturns(acceptEULAErr)
		fakePivnetClient.GetProductFilesReturns(productFilesResponse, getProductFilesErr)

		fakePivnetClient.GetProductFileStub = func(
			productSlug string,
			releaseID int,
			productFileID int,
		) (pivnet.ProductFile, error) {
			if getProductFileErr != nil {
				return pivnet.ProductFile{}, getProductFileErr
			}

			for _, p := range productFiles {
				if p.ID == productFileID {
					return p, nil
				}
			}

			Fail("unexpected productFileID: %d", productFileID)
			return pivnet.ProductFile{}, nil
		}

		fakeFilter.DownloadLinksByGlobReturns(map[string]string{}, downloadLinksByGlobErr)

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		inCommand = in.NewInCommand(ginkgoLogger, downloadDir, fakePivnetClient, fakeFilter, fakeDownloader)
	})

	AfterEach(func() {
		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes a version file with the downloaded version and etag", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "version")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(versionContents)).To(Equal(versionWithETag))
	})

	var validateProductFilesMetadata = func(
		writtenMetadata metadata.Metadata,
		productFiles []pivnet.ProductFile,
	) {
		Expect(writtenMetadata.ProductFiles).To(HaveLen(len(productFiles)))
		for i, p := range productFiles {
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

	It("writes a metadata file in yaml format", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "metadata.yaml")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())

		var writtenMetadata metadata.Metadata
		err = yaml.Unmarshal(versionContents, &writtenMetadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(writtenMetadata.Release).NotTo(BeNil())
		Expect(writtenMetadata.Release.Version).To(Equal(productVersion))

		validateProductFilesMetadata(writtenMetadata, productFiles)
	})

	It("writes a metadata file in json format", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "metadata.json")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())

		var writtenMetadata metadata.Metadata
		err = json.Unmarshal(versionContents, &writtenMetadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(writtenMetadata.Release).NotTo(BeNil())
		Expect(writtenMetadata.Release.Version).To(Equal(productVersion))

		validateProductFilesMetadata(writtenMetadata, productFiles)
	})

	It("does not download any of the files in the specified release", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		files, err := ioutil.ReadDir(downloadDir)
		Expect(err).ShouldNot(HaveOccurred())

		// the version and metadata files will always exist
		Expect(len(files)).To(Equal(3))
		Expect(files[0].Name()).To(Equal("metadata.json"))
		Expect(files[1].Name()).To(Equal("metadata.yaml"))
		Expect(files[2].Name()).To(Equal("version"))
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

	Context("when creating download dir fails", func() {
		BeforeEach(func() {
			downloadDir = "/not/a/real/dir"
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when getting release returns error", func() {
		BeforeEach(func() {
			getReleaseErr = fmt.Errorf("some error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getReleaseErr))
		})
	})

	Context("when accepting EULA returns error", func() {
		BeforeEach(func() {
			acceptEULAErr = fmt.Errorf("some error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(acceptEULAErr))
		})
	})

	Context("when getting product files returns error", func() {
		BeforeEach(func() {
			getProductFilesErr = fmt.Errorf("some error")
		})

		It("returns error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getProductFilesErr))
		})
	})

	Context("when getting a product file returns error", func() {
		BeforeEach(func() {
			getProductFileErr = fmt.Errorf("some error")
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

		Context("when filtering download links returns error", func() {
			BeforeEach(func() {
				downloadLinksByGlobErr = fmt.Errorf("some error")
			})

			It("returns error", func() {
				_, err := inCommand.Run(inRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(downloadLinksByGlobErr))
			})
		})
	})
})
