package downloader_test

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/downloader"
	"github.com/pivotal-cf/pivnet-resource/downloader/downloaderfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader", func() {
	var (
		fakeClient *downloaderfakes.FakeClient
		d          *downloader.Downloader
		dir        string
		fakeLogger logger.Logger
	)

	BeforeEach(func() {
		fakeClient = &downloaderfakes.FakeClient{}

		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		var err error
		dir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		d = downloader.NewDownloader(fakeClient, dir, fakeLogger, GinkgoWriter)
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Download", func() {
		var (
			productSlug  string
			releaseID    int
			productFiles []pivnet.ProductFile
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			releaseID = 1234

			productFiles = []pivnet.ProductFile{
				{
					ID:           1337,
					Name:         "pf-0",
					AWSObjectKey: "bucket/path/file-0",
				},
				{
					ID:           1234,
					Name:         "pf-1",
					AWSObjectKey: "bucket/path/file-1",
				},
				{
					ID:           1886,
					Name:         "pf-2",
					AWSObjectKey: "bucket/path/file-2",
				},
			}
		})

		It("downloads all of the product files", func() {
			filepaths, err := d.Download(productFiles, productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.DownloadProductFileCallCount()).To(Equal(3))

			f, slug, relID, productFileID, w := fakeClient.DownloadProductFileArgsForCall(0)
			Expect(f.Name()).To(BeAnExistingFile())
			Expect(slug).To(Equal(productSlug))
			Expect(relID).To(Equal(releaseID))
			Expect(productFileID).To(Equal(productFiles[0].ID))
			Expect(w).To(Equal(GinkgoWriter))

			f, slug, relID, productFileID, w = fakeClient.DownloadProductFileArgsForCall(1)
			Expect(f.Name()).To(BeAnExistingFile())
			Expect(slug).To(Equal(productSlug))
			Expect(relID).To(Equal(releaseID))
			Expect(productFileID).To(Equal(productFiles[1].ID))
			Expect(w).To(Equal(GinkgoWriter))

			f, slug, relID, productFileID, w = fakeClient.DownloadProductFileArgsForCall(2)
			Expect(f.Name()).To(BeAnExistingFile())
			Expect(slug).To(Equal(productSlug))
			Expect(relID).To(Equal(releaseID))
			Expect(productFileID).To(Equal(productFiles[2].ID))
			Expect(w).To(Equal(GinkgoWriter))

			Expect(len(filepaths)).To(Equal(3))

			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-0")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-1")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-2")))
		})

		Context("when the pivnet client returns an error", func() {
			BeforeEach(func() {
				productFiles = []pivnet.ProductFile{
					{
						Name:         "pf-0",
						AWSObjectKey: "bucket/path/file-0",
					},
				}
			})

			Context("when the pivnet client returns other errors", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("download file error")
					fakeClient.DownloadProductFileReturns(expectedErr)
				})

				It("raises an error", func() {
					_, err := d.Download(productFiles, productSlug, releaseID)

					Expect(err).Should(HaveOccurred())
					Expect(err).To(Equal(expectedErr))
					Expect(fakeClient.DownloadProductFileCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the directory does not already exist", func() {
			BeforeEach(func() {
				dir = filepath.Join(dir, "sub_directory")
			})

			It("creates the directory", func() {
				_, err := d.Download(productFiles, productSlug, releaseID)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Open(dir)
				Expect(err).ShouldNot(HaveOccurred())
			})

			Context("when creating the directory fails", func() {
				BeforeEach(func() {
					dir = "/proc/nope"
				})

				It("returns an error", func() {
					_, err := d.Download(productFiles, productSlug, releaseID)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when it fails to create a file", func() {
			BeforeEach(func() {
				productFiles[0].AWSObjectKey = "/"
			})

			It("returns an error", func() {
				_, err := d.Download(productFiles, productSlug, releaseID)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
