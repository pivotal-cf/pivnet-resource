package uploader_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	uploader_fakes "github.com/pivotal-cf-experimental/pivnet-resource/uploader/fakes"
)

var _ = Describe("Uploader", func() {
	Describe("ExactGlobs", func() {
		var (
			l              logger.Logger
			fakeTransport  *uploader_fakes.FakeTransport
			uploaderConfig uploader.Config
			uploaderClient uploader.Client

			tempDir    string
			myFilesDir string

			filepathPrefix = "Some-Filepath-Prefix"
		)

		BeforeEach(func() {
			l = logger.NewLogger(GinkgoWriter)
			fakeTransport = &uploader_fakes.FakeTransport{}

			var err error
			tempDir, err = ioutil.TempDir("", "pivnet-resource")
			Expect(err).NotTo(HaveOccurred())

			myFilesDir = filepath.Join(tempDir, "my_files")
			err = os.Mkdir(myFilesDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Create(filepath.Join(myFilesDir, "file-0"))
			Expect(err).NotTo(HaveOccurred())

			uploaderConfig = uploader.Config{
				FilepathPrefix: filepathPrefix,
				FileGlob:       "my_files/*",
				Transport:      fakeTransport,
				SourcesDir:     tempDir,
				Logger:         l,
			}

			uploaderClient = uploader.NewClient(uploaderConfig)
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no files match the fileglob", func() {
			BeforeEach(func() {
				uploaderConfig.FileGlob = "this-will-match-nothing"
				uploaderClient = uploader.NewClient(uploaderConfig)
			})

			It("returns an error", func() {
				_, err := uploaderClient.ExactGlobs()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("no matches"))
			})
		})

		Context("when multiple files match the fileglob", func() {
			BeforeEach(func() {
				_, err := os.Create(filepath.Join(myFilesDir, "file-1"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a map of filenames to remote paths", func() {
				filenamePaths, err := uploaderClient.ExactGlobs()
				Expect(err).NotTo(HaveOccurred())

				Expect(len(filenamePaths)).To(Equal(2))

				Expect(filenamePaths[0]).To(Equal("my_files/file-0"))
				Expect(filenamePaths[1]).To(Equal("my_files/file-1"))
			})
		})
	})

	Describe("UploadFile", func() {
		var (
			l              logger.Logger
			fakeTransport  *uploader_fakes.FakeTransport
			uploaderConfig uploader.Config
			uploaderClient uploader.Client

			tempDir    string
			myFilesDir string

			filepathPrefix = "Some-Filepath-Prefix"
		)

		BeforeEach(func() {
			l = logger.NewLogger(GinkgoWriter)
			fakeTransport = &uploader_fakes.FakeTransport{}

			var err error
			tempDir, err = ioutil.TempDir("", "pivnet-resource")
			Expect(err).NotTo(HaveOccurred())

			myFilesDir = filepath.Join(tempDir, "my_files")
			err = os.Mkdir(myFilesDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Create(filepath.Join(myFilesDir, "file-0"))
			Expect(err).NotTo(HaveOccurred())

			uploaderConfig = uploader.Config{
				FilepathPrefix: filepathPrefix,
				FileGlob:       "my_files/*",
				Transport:      fakeTransport,
				SourcesDir:     tempDir,
				Logger:         l,
			}

			uploaderClient = uploader.NewClient(uploaderConfig)
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the file exists", func() {
			BeforeEach(func() {
				_, err := os.Create(filepath.Join(myFilesDir, "file-0"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("invokes the transport", func() {
				_, err := uploaderClient.UploadFile("my_files/file-0")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTransport.UploadCallCount()).To(Equal(1))

				glob0, _, _ := fakeTransport.UploadArgsForCall(0)
				Expect(glob0).To(Equal("my_files/file-0"))
			})

			It("returns a map of filenames to remote paths", func() {
				remotePath, err := uploaderClient.UploadFile("my_files/file-0")
				Expect(err).NotTo(HaveOccurred())

				Expect(remotePath).To(Equal(
					fmt.Sprintf("product_files/%s/file-0", filepathPrefix)))
			})
		})

		Context("when the transport exits with error", func() {
			BeforeEach(func() {
				fakeTransport.UploadReturns(errors.New("some error"))
			})

			It("propagates errors", func() {
				_, err := uploaderClient.UploadFile("foo")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("some error"))
			})
		})

		Context("when the glob is empty", func() {
			It("returns an error", func() {
				_, err := uploaderClient.UploadFile("")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("glob"))
			})
		})
	})
})
