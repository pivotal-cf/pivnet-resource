package uploader_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	uploader_fakes "github.com/pivotal-cf-experimental/pivnet-resource/uploader/fakes"
)

var _ = Describe("Uploader", func() {
	var (
		debugWriter    io.Writer
		fakeTransport  *uploader_fakes.FakeTransport
		uploaderConfig uploader.Config
		uploaderClient uploader.Client

		tempDir    string
		myFilesDir string
	)

	BeforeEach(func() {
		debugWriter = GinkgoWriter
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
			FileGlob:    "my_files/*",
			Transport:   fakeTransport,
			SourcesDir:  tempDir,
			DebugWriter: debugWriter,
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
			err := uploaderClient.Upload()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("no matches"))
		})
	})

	Context("when multiple files match the fileglob", func() {
		BeforeEach(func() {
			_, err := os.Create(filepath.Join(myFilesDir, "file-1"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("invokes the transport once per file with a separate glob per file", func() {
			err := uploaderClient.Upload()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTransport.UploadCallCount()).To(Equal(2))

			glob0, _, _ := fakeTransport.UploadArgsForCall(0)
			Expect(glob0).To(Equal("my_files/file-0"))

			glob1, _, _ := fakeTransport.UploadArgsForCall(1)
			Expect(glob1).To(Equal("my_files/file-1"))
		})
	})

	Context("when the transport exits with error", func() {
		BeforeEach(func() {
			fakeTransport.UploadReturns(errors.New("some error"))
		})

		It("propagates errors", func() {
			err := uploaderClient.Upload()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})
})
