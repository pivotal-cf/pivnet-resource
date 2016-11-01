package downloader_test

import (
	"errors"
	"io"
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
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Downloader", func() {
	var (
		fakeClient *downloaderfakes.FakeClient
		d          *downloader.Downloader
		server     *ghttp.Server
		apiAddress string
		dir        string
		fakeLogger logger.Logger
	)

	BeforeEach(func() {
		fakeClient = &downloaderfakes.FakeClient{}

		server = ghttp.NewServer()
		apiAddress = server.URL()
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		var err error
		dir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		d = downloader.NewDownloader(fakeClient, dir, fakeLogger)
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
					Name:         "pf-0",
					AWSObjectKey: "bucket/path/file-0",
				},
				{
					Name:         "pf-1",
					AWSObjectKey: "bucket/path/file-1",
				},
				{
					Name:         "pf-2",
					AWSObjectKey: "bucket/path/file-2",
				},
			}
		})

		It("returns a list of (full) filepaths", func() {
			filepaths, err := d.Download(productFiles, productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(filepaths)).To(Equal(3))

			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-0")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-1")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-2")))
		})

		Context("when the pivnet client returns an error", func() {
			var (
				expectedErr *downloaderfakes.FakeNetError
			)

			BeforeEach(func() {
				productFiles = []pivnet.ProductFile{
					{
						Name:         "pf-0",
						AWSObjectKey: "bucket/path/file-0",
					},
				}
			})

			Context("when the pivnet client returns a network error", func() {
				BeforeEach(func() {
					expectedErr = &downloaderfakes.FakeNetError{}
					fakeClient.DownloadProductFileReturns(expectedErr)
				})

				Context("when the network error is temporary", func() {
					BeforeEach(func() {
						expectedErr.TemporaryReturns(true)
					})

					It("attempts three downloads", func() {
						_, err := d.Download(productFiles, productSlug, releaseID)

						Expect(err).Should(HaveOccurred())
						Expect(err).To(Equal(expectedErr))
						Expect(fakeClient.DownloadProductFileCallCount()).To(Equal(3))
					})

					Context("when the download succeeds the second or third time", func() {
						BeforeEach(func() {
							remainingFailures := 1

							fakeClient.DownloadProductFileStub = func(w io.Writer, s string, r int, p int) error {
								if remainingFailures > 0 {
									remainingFailures--
									return expectedErr
								}

								return nil
							}
						})

						It("does not throw an error", func() {
							_, err := d.Download(productFiles, productSlug, releaseID)

							Expect(err).ShouldNot(HaveOccurred())
							Expect(fakeClient.DownloadProductFileCallCount()).To(Equal(2))
						})
					})
				})

				Context("when the network error is not temporary", func() {
					It("raises an error", func() {
						_, err := d.Download(productFiles, productSlug, releaseID)

						Expect(err).Should(HaveOccurred())
						Expect(err).To(Equal(expectedErr))
						Expect(fakeClient.DownloadProductFileCallCount()).To(Equal(1))
					})
				})
			})

			Context("when it encounters other errors", func() {
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
