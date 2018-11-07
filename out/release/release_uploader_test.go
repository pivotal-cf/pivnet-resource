package release_test

import (
	"errors"
	"log"
	"time"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseUploader", func() {
	var (
		fakeLogger logger.Logger

		s3Client      *releasefakes.S3Client
		uploadClient  *releasefakes.UploadClient
		sha256Summer  *releasefakes.Sha256Summer
		md5Summer     *releasefakes.Md5Summer
		pivnetRelease pivnet.Release
		uploader      release.ReleaseUploader
		asyncTimeout  time.Duration
		pollFrequency time.Duration

		productSlug string

		mdata metadata.Metadata

		existingProductFiles []pivnet.ProductFile
		actualSHA256Sum      string
		actualMD5Sum         string
		newAWSObjectKey      string
		productFileTransferStatus string

		existingProductFilesErr error
		createProductFileErr    error
		uploadFileErr           error
		computeAWSObjectKeyError  error
		sha256SumFileErr        error
		md5SumFileErr           error
		productFileErr          error
	)

	BeforeEach(func() {
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		s3Client = &releasefakes.S3Client{}
		uploadClient = &releasefakes.UploadClient{}
		sha256Summer = &releasefakes.Sha256Summer{}
		md5Summer = &releasefakes.Md5Summer{}

		productSlug = "some-product-slug"

		asyncTimeout = 450 * time.Millisecond
		pollFrequency = 15 * time.Millisecond

		pivnetRelease = pivnet.Release{
			ID:      1111,
			Version: "some-release-version",
		}

		mdata = metadata.Metadata{
			ProductFiles: []metadata.ProductFile{
				{
					File:               "some/file",
					Description:        "a description",
					UploadAs:           "upload as",
					FileType:           "something",
					DocsURL:            "some-docs-url",
					SystemRequirements: []string{"req1", "req2"},
					Platforms:          []string{"Linux"},
					IncludedFiles:      []string{"include1", "include2"},
				},
			},
		}

		existingProductFiles = []pivnet.ProductFile{
			{
				ID:           1234,
				AWSObjectKey: "some-existing-aws-object-key",
				Name: mdata.ProductFiles[0].UploadAs,
			},
		}

		actualSHA256Sum = "madeupsha256"
		actualMD5Sum = "madeupmd5"
		newAWSObjectKey = "s3-remote-path"
		productFileTransferStatus = "complete"

		existingProductFilesErr = nil
		createProductFileErr = nil
		uploadFileErr = nil
		computeAWSObjectKeyError = nil
		sha256SumFileErr = nil
		md5SumFileErr = nil
		productFileErr = nil
	})

	JustBeforeEach(func() {
		uploader = release.NewReleaseUploader(
			s3Client,
			uploadClient,
			fakeLogger,
			sha256Summer,
			md5Summer,
			mdata,
			"/some/sources/dir",
			productSlug,
			asyncTimeout,
			pollFrequency,
		)

		sha256Summer.SumFileReturns(actualSHA256Sum, sha256SumFileErr)
		md5Summer.SumFileReturns(actualMD5Sum, md5SumFileErr)
		s3Client.UploadFileReturns(uploadFileErr)
		s3Client.ComputeAWSObjectKeyReturns(newAWSObjectKey, "", computeAWSObjectKeyError)
		uploadClient.CreateProductFileReturns(pivnet.ProductFile{ID: 13367}, createProductFileErr)
		uploadClient.ProductFilesReturns(existingProductFiles, existingProductFilesErr)

		invokeCount := 0
		uploadClient.ProductFileStub = func(string, int) (pivnet.ProductFile, error) {
			if productFileErr != nil {
				return pivnet.ProductFile{}, productFileErr
			}

			productFile := existingProductFiles[0]

			invokeCount += 1

			if invokeCount == 1 {
				productFile.FileTransferStatus = "in_progress"
				return productFile, nil
			}

			productFile.FileTransferStatus = productFileTransferStatus
			return productFile, nil
		}
	})

	Describe("Upload", func() {
		It("uploads a release to s3 and adds metadata to pivnet", func() {
			err := uploader.Upload(pivnetRelease, []string{"some/file"})
			Expect(err).NotTo(HaveOccurred())

			Expect(sha256Summer.SumFileArgsForCall(0)).To(Equal("/some/sources/dir/some/file"))
			Expect(md5Summer.SumFileArgsForCall(0)).To(Equal("/some/sources/dir/some/file"))
			Expect(s3Client.UploadFileArgsForCall(0)).To(Equal("some/file"))

			Expect(uploadClient.CreateProductFileArgsForCall(0)).To(Equal(pivnet.CreateProductFileConfig{
				ProductSlug:        productSlug,
				AWSObjectKey:       newAWSObjectKey,
				SHA256:             actualSHA256Sum,
				MD5:                actualMD5Sum,
				FileVersion:        pivnetRelease.Version,
				Name:               mdata.ProductFiles[0].UploadAs,
				Description:        mdata.ProductFiles[0].Description,
				FileType:           mdata.ProductFiles[0].FileType,
				DocsURL:            mdata.ProductFiles[0].DocsURL,
				SystemRequirements: mdata.ProductFiles[0].SystemRequirements,
				Platforms:          mdata.ProductFiles[0].Platforms,
				IncludedFiles:      mdata.ProductFiles[0].IncludedFiles,
			}))

			invokedProductSlug, releaseID, productFileID := uploadClient.AddProductFileArgsForCall(0)
			Expect(invokedProductSlug).To(Equal(productSlug))
			Expect(releaseID).To(Equal(1111))
			Expect(productFileID).To(Equal(13367))
		})

		Context("when a product file already exists with AWSObjectKey", func() {
			BeforeEach(func() {
				newAWSObjectKey = existingProductFiles[0].AWSObjectKey
			})
			Context("when the files have the same content", func() {
				BeforeEach(func() {
					existingProductFiles[0].SHA256 = actualSHA256Sum
					existingProductFiles[0].MD5 = actualMD5Sum
				})

				It("should not re-upload the file to S3", func() {
					err := uploader.Upload(pivnetRelease, []string{"some/file"})
					Expect(err).NotTo(HaveOccurred())
					Expect(s3Client.UploadFileCallCount()).To(Equal(0))
				})

				It("should NOT delete the product and associate the existing product file", func() {
					err := uploader.Upload(pivnetRelease, []string{"some/file"})
					Expect(err).NotTo(HaveOccurred())
					Expect(uploadClient.DeleteProductFileCallCount()).To(Equal(0))
					Expect(uploadClient.CreateProductFileCallCount()).To(Equal(0))
					Expect(uploadClient.AddProductFileCallCount()).To(Equal(1))
				})

				Context("when the UploadAs metadata differes from the product file name in pivnet", func() {
					BeforeEach(func() {
						existingProductFiles[0].Name = "different_name"
					})

					It("should display an error", func() {
						err := uploader.Upload(pivnetRelease, []string{"some/file"})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("A file with the same name was found"))
						Expect(err.Error()).To(ContainSubstring(existingProductFiles[0].Name))
						Expect(err.Error()).To(ContainSubstring("UploadAs metadata value: Unknown"))
						Expect(err.Error()).To(ContainSubstring(existingProductFiles[0].AWSObjectKey))
					})
				})
			})
			Context("when the files have different content", func() {
				It("should display error message", func() {
					err := uploader.Upload(pivnetRelease, []string{"some/file"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("already exists on S3"))
				})
			})
		})

		Context("when the file sha256 cannot be computed", func() {
			BeforeEach(func() {
				sha256SumFileErr = errors.New("sha256 error")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(MatchError(errors.New("sha256 error")))
			})
		})

		Context("when the file md5 cannot be computed", func() {
			BeforeEach(func() {
				md5SumFileErr = errors.New("md5 error")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(MatchError(errors.New("md5 error")))
			})
		})

		Context("when computing the AWS Object Key fails", func() {
			BeforeEach(func() {
				computeAWSObjectKeyError = errors.New("AWS object key generation fails")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(Equal(computeAWSObjectKeyError))
			})
		})

		Context("when the s3 upload fails", func() {
			BeforeEach(func() {
				uploadFileErr = errors.New("s3 failed")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(Equal(uploadFileErr))
			})
		})

		Context("when pivnet fails to find a product", func() {
			BeforeEach(func() {
				createProductFileErr = errors.New("some product files error")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(Equal(createProductFileErr))
			})
		})

		Context("when pivnet fails to get existing product files", func() {
			BeforeEach(func() {
				existingProductFilesErr = errors.New("some product files error")
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(Equal(existingProductFilesErr))
			})
		})

		Context("when pivnet cannot add a product file", func() {
			BeforeEach(func() {
				uploadClient.AddProductFileReturns(errors.New("error adding product"))
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(MatchError(errors.New("error adding product")))
			})
		})

		Context("when polling for the product file returns an error", func() {
			Context("when getting product file from pivnet fails", func() {
				BeforeEach(func() {
					productFileErr = errors.New("product file error")
				})

				It("returns an error", func() {
					err := uploader.Upload(pivnetRelease, []string{""})
					Expect(err).To(Equal(productFileErr))
				})
			})

			Context("When file_transfer_status returns an error", func() {
				It("returns a file_transfer_status that is a failed_sha256_check", func() {
					productFileTransferStatus = "failed_sha256_check"
					err := uploader.Upload(pivnetRelease, []string{""})
					Expect(err).To(MatchError(errors.New("failed_sha256_check")))
				})

				It("returns a file_transfer_status that is a failed_md5_check", func() {
					productFileTransferStatus = "failed_md5_check"
					err := uploader.Upload(pivnetRelease, []string{""})
					Expect(err).To(MatchError(errors.New("failed_md5_check")))
				})
			})
		})

		Context("when polling for the product file times out", func() {
			BeforeEach(func() {
				asyncTimeout = pollFrequency / 2
			})

			It("returns an error", func() {
				err := uploader.Upload(pivnetRelease, []string{""})
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("timed out"))
			})
		})

	})

})
