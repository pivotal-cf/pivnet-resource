package downloader_test

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
	"github.com/pivotal-cf-experimental/pivnet-resource/downloader/downloaderfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Downloader", func() {
	var (
		fakeExtendedClient *downloaderfakes.FakeExtendedClient
		d                  *downloader.Downloader
		server             *ghttp.Server
		apiAddress         string
		dir                string
		logger             *log.Logger
	)

	BeforeEach(func() {
		fakeExtendedClient = &downloaderfakes.FakeExtendedClient{}

		server = ghttp.NewServer()
		apiAddress = server.URL()
		logger = log.New(ioutil.Discard, "doesn't matter", 0)

		var err error
		dir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		d = downloader.NewDownloader(fakeExtendedClient, dir, logger)
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Download", func() {
		It("returns a list of (full) filepaths", func() {
			fileNames := map[string]string{
				"file-0": "/file-0",
				"file-1": "/file-1",
				"file-2": "/file-2",
			}

			filepaths, err := d.Download(fileNames)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(filepaths)).To(Equal(3))

			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-0")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-1")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-2")))
		})

		Context("when it fails to make a request", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("download file error")
				fakeExtendedClient.DownloadFileReturns(expectedErr)
			})

			It("raises an error", func() {
				_, err := d.Download(map[string]string{"^731drop": "&h%%%%"})

				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when the directory does not already exist", func() {
			BeforeEach(func() {
				dir = filepath.Join(dir, "sub_directory")
			})

			It("creates the directory", func() {
				_, err := d.Download(nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Open(dir)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when it fails to create a file", func() {
			It("returns an error", func() {
				_, err := d.Download(map[string]string{"/": ""})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
