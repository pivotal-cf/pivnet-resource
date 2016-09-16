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

			filepathPrefix string
		)

		BeforeEach(func() {
			fakeTransport = &uploaderfakes.FakeTransport{}

			filepathPrefix = "Some-Filepath-Prefix"

			var err error
			tempDir, err = ioutil.TempDir("", "pivnet-resource")
			Expect(err).NotTo(HaveOccurred())

			myFilesDir = filepath.Join(tempDir, "my_files")
			err = os.Mkdir(myFilesDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Create(filepath.Join(myFilesDir, "file-0"))
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
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

		It("invokes the transport", func() {
			_, err := uploaderClient.UploadFile("my_files/file-0")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTransport.UploadCallCount()).To(Equal(1))

			glob0, remoteDir, sourcesDir := fakeTransport.UploadArgsForCall(0)
			Expect(glob0).To(Equal("my_files/file-0"))
			Expect(remoteDir).To(Equal(fmt.Sprintf("product_files/%s/", filepathPrefix)))
			Expect(sourcesDir).To(Equal(tempDir))
		})

		It("returns a map of filenames to remote paths", func() {
			remotePath, err := uploaderClient.UploadFile("my_files/file-0")
			Expect(err).NotTo(HaveOccurred())

			Expect(remotePath).To(Equal(
				fmt.Sprintf("product_files/%s/file-0", filepathPrefix)))
		})

		Context("when the filepathPrefix begins with 'product_files'", func() {
			var (
				oldFilepathPrefix string
			)

			BeforeEach(func() {
				oldFilepathPrefix = filepathPrefix
				filepathPrefix = fmt.Sprintf("product_files/%s", filepathPrefix)
			})

			It("invokes the transport with 'product_files'", func() {
				_, err := uploaderClient.UploadFile("my_files/file-0")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTransport.UploadCallCount()).To(Equal(1))

				glob0, remoteDir, sourcesDir := fakeTransport.UploadArgsForCall(0)
				Expect(glob0).To(Equal("my_files/file-0"))
				Expect(remoteDir).To(Equal(fmt.Sprintf("product_files/%s/", oldFilepathPrefix)))
				Expect(sourcesDir).To(Equal(tempDir))
			})
		})

		Context("when the filepathPrefix begins with 'product-files'", func() {
			var (
				oldFilepathPrefix string
			)

			BeforeEach(func() {
				oldFilepathPrefix = filepathPrefix
				filepathPrefix = fmt.Sprintf("product-files/%s", filepathPrefix)
			})

			It("invokes the transport with 'product-files'", func() {
				_, err := uploaderClient.UploadFile("my_files/file-0")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeTransport.UploadCallCount()).To(Equal(1))

				glob0, remoteDir, sourcesDir := fakeTransport.UploadArgsForCall(0)
				Expect(glob0).To(Equal("my_files/file-0"))
				Expect(remoteDir).To(Equal(fmt.Sprintf("product-files/%s/", oldFilepathPrefix)))
				Expect(sourcesDir).To(Equal(tempDir))
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
