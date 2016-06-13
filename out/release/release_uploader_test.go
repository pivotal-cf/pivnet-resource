package release_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release/releasefakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseUploader", func() {
	var (
		s3Client      *releasefakes.S3Client
		logging       *releasefakes.Logging
		uploadClient  *releasefakes.UploadClient
		md5Summer     *releasefakes.Md5Summer
		pivnetRelease pivnet.Release
		uploader      release.ReleaseUploader
	)

	BeforeEach(func() {
		s3Client = &releasefakes.S3Client{}
		logging = &releasefakes.Logging{}
		uploadClient = &releasefakes.UploadClient{}
		md5Summer = &releasefakes.Md5Summer{}

		pivnetRelease = pivnet.Release{
			ID:      1111,
			Version: "some-release-version",
		}

		meta := metadata.Metadata{
			ProductFiles: []metadata.ProductFile{
				{
					File:        "some/file",
					Description: "a description",
					UploadAs:    "a file",
				},
			},
		}

		uploader = release.NewReleaseUploader(s3Client,
			uploadClient,
			logging,
			md5Summer,
			meta,
			false,
			"/some/sources/dir",
			"some-product-slug",
		)
	})

	Describe("Upload", func() {
		Context("when the upload is not skipped", func() {
			BeforeEach(func() {
				md5Summer.SumFileReturns("madeupmd5", nil)
				s3Client.UploadFileReturns("s3-remote-path", nil)
				uploadClient.CreateProductFileReturns(pivnet.ProductFile{ID: 13367}, nil)
				uploadClient.FindProductForSlugReturns(pivnet.Product{ID: 7777}, nil)
			})

			It("uploads a release to s3 and adds metadata to pivnet", func() {
				err := uploader.Upload(pivnetRelease, []string{"some/file"})
				Expect(err).NotTo(HaveOccurred())

				Expect(md5Summer.SumFileArgsForCall(0)).To(Equal("/some/sources/dir/some/file"))
				Expect(s3Client.UploadFileArgsForCall(0)).To(Equal("some/file"))
				Expect(uploadClient.FindProductForSlugArgsForCall(0)).To(Equal("some-product-slug"))

				message, types := logging.DebugfArgsForCall(0)
				Expect(message).To(Equal("exact glob '%s' matches metadata file: '%s'\n"))
				Expect(types[0].(string)).To(Equal("some/file"))
				Expect(types[1].(string)).To(Equal("some/file"))

				message, types = logging.DebugfArgsForCall(1)
				Expect(message).To(Equal("upload_as provided for exact glob: '%s' - uploading to remote filename: '%s' instead\n"))
				Expect(types[0].(string)).To(Equal("some/file"))
				Expect(types[1].(string)).To(Equal("a file"))

				message, types = logging.DebugfArgsForCall(2)
				Expect(message).To(Equal("Creating product file: {product_slug: %s, filename: %s, aws_object_key: %s, file_version: %s, description: %s}\n"))
				Expect(types[0].(string)).To(Equal("some-product-slug"))
				Expect(types[1].(string)).To(Equal("a file"))
				Expect(types[2].(string)).To(Equal("s3-remote-path"))
				Expect(types[3].(string)).To(Equal("some-release-version"))
				Expect(types[4].(string)).To(Equal("a description"))

				Expect(uploadClient.CreateProductFileArgsForCall(0)).To(Equal(pivnet.CreateProductFileConfig{
					ProductSlug:  "some-product-slug",
					Name:         "a file",
					AWSObjectKey: "s3-remote-path",
					FileVersion:  "some-release-version",
					MD5:          "madeupmd5",
					Description:  "a description",
				}))

				message, types = logging.DebugfArgsForCall(3)
				Expect(message).To(Equal("Adding product file: {product_slug: %s, product_id: %d, filename: %s, product_file_id: %d, release_id: %d}\n"))
				Expect(types[0].(string)).To(Equal("some-product-slug"))
				Expect(types[1].(int)).To(Equal(7777))
				Expect(types[2].(string)).To(Equal("file"))
				Expect(types[3].(int)).To(Equal(13367))
				Expect(types[4].(int)).To(Equal(1111))

				productID, releaseID, productFileID := uploadClient.AddProductFileArgsForCall(0)
				Expect(productID).To(Equal(7777))
				Expect(releaseID).To(Equal(1111))
				Expect(productFileID).To(Equal(13367))
			})

			Context("when the glob do not match", func() {
				It("logs a message", func() {
					err := uploader.Upload(pivnetRelease, []string{"/broken/glob"})
					Expect(err).NotTo(HaveOccurred())

					Expect(logging.DebugfCallCount()).To(Equal(3))

					message, types := logging.DebugfArgsForCall(0)
					Expect(message).To(Equal("exact glob %s does not match metadata file: %s\n"))
					Expect(types[0].(string)).To(Equal("/broken/glob"))
					Expect(types[1].(string)).To(Equal("some/file"))
				})
			})

			Context("when an error occurs", func() {
				Context("when the file md5 cannot be computed", func() {
					BeforeEach(func() {
						md5Summer.SumFileReturns("", errors.New("md5 error"))
					})

					It("returns an error", func() {
						err := uploader.Upload(pivnetRelease, []string{""})
						Expect(err).To(MatchError(errors.New("md5 error")))
					})
				})

				Context("when the s3 upload fails", func() {
					BeforeEach(func() {
						s3Client.UploadFileReturns("", errors.New("s3 failed"))
					})

					It("returns an error", func() {
						err := uploader.Upload(pivnetRelease, []string{""})
						Expect(err).To(MatchError(errors.New("s3 failed")))
					})
				})

				Context("when pivnet fails to find a product", func() {
					BeforeEach(func() {
						uploadClient.CreateProductFileReturns(pivnet.ProductFile{}, errors.New("pivnet product blew up"))
					})

					It("returns an error", func() {
						err := uploader.Upload(pivnetRelease, []string{""})
						Expect(err).To(MatchError("pivnet product blew up"))
					})
				})

				Context("when pivnet cannot create a product file", func() {
					BeforeEach(func() {
						uploadClient.FindProductForSlugReturns(pivnet.Product{}, errors.New("cannot create product file"))
					})

					It("returns an error", func() {
						err := uploader.Upload(pivnetRelease, []string{""})
						Expect(err).To(MatchError(errors.New("cannot create product file")))
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
			})
		})

		Context("when the upload is skipped", func() {
			BeforeEach(func() {
				uploader = release.NewReleaseUploader(nil, nil, logging, nil, metadata.Metadata{}, true, "", "")
			})

			It("logs a message", func() {
				err := uploader.Upload(pivnetRelease, []string{"/some/file"})
				Expect(err).NotTo(HaveOccurred())

				message, _ := logging.DebugfArgsForCall(0)
				Expect(message).To(Equal("File glob and s3_filepath_prefix not provided - skipping upload to s3"))

				Expect(s3Client.UploadFileCallCount()).To(BeZero())
			})
		})
	})
})
