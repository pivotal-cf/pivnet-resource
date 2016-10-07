package release_test

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseUploader", func() {
	var (
		s3Client      *releasefakes.S3Client
		logging       *log.Logger
		uploadClient  *releasefakes.UploadClient
		md5Summer     *releasefakes.Md5Summer
		pivnetRelease pivnet.Release
		uploader      release.ReleaseUploader
	)

	BeforeEach(func() {
		s3Client = &releasefakes.S3Client{}
		logging = log.New(ioutil.Discard, "it doesn't matter", 0)
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
					FileType:    "something",
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
			})

			It("uploads a release to s3 and adds metadata to pivnet", func() {
				err := uploader.Upload(pivnetRelease, []string{"some/file"})
				Expect(err).NotTo(HaveOccurred())

				Expect(md5Summer.SumFileArgsForCall(0)).To(Equal("/some/sources/dir/some/file"))
				Expect(s3Client.UploadFileArgsForCall(0)).To(Equal("some/file"))

				Expect(uploadClient.CreateProductFileArgsForCall(0)).To(Equal(pivnet.CreateProductFileConfig{
					ProductSlug:  "some-product-slug",
					Name:         "a file",
					AWSObjectKey: "s3-remote-path",
					FileVersion:  "some-release-version",
					MD5:          "madeupmd5",
					Description:  "a description",
					FileType:     "something",
				}))

				productSlug, releaseID, productFileID := uploadClient.AddProductFileArgsForCall(0)
				Expect(productSlug).To(Equal("some-product-slug"))
				Expect(releaseID).To(Equal(1111))
				Expect(productFileID).To(Equal(13367))
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
	})
})
