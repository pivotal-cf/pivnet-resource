package uploader_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/uploader"
	"github.com/pivotal-cf/pivnet-resource/uploader/uploaderfakes"
)

var _ = Describe("Uploader", func() {
	Describe("UploadFile", func() {
		var (
			fakeTransport  *uploaderfakes.FakeTransport
			uploaderConfig uploader.Config
			uploaderClient *uploader.Client

			tempDir    string
			myFilesDir string

			filepathPrefix = "Some-Filepath-Prefix"
		)

		BeforeEach(func() {
			fakeTransport = &uploaderfakes.FakeTransport{}

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
				Transport:      fakeTransport,
				SourcesDir:     tempDir,
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
					fmt.Sprintf("%s/file-0", filepathPrefix)))
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
