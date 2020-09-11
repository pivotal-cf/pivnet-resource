package uploader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/v2/uploader"
	"github.com/pivotal-cf/pivnet-resource/v2/uploader/uploaderfakes"
	"errors"
	"fmt"
)

var _ = Describe("Uploader", func() {
	var (
		fakeTransport  *uploaderfakes.FakeTransport
		uploaderConfig uploader.Config
		uploaderClient *uploader.Client

		exactGlob string
		tempDir    string

		filepathPrefix string
	)

	BeforeEach(func() {
		fakeTransport = &uploaderfakes.FakeTransport{}

		filepathPrefix = "product-files/my-product-slug"
		exactGlob = "my-product-file"
		tempDir = "my/temp/dir"
	})

	JustBeforeEach(func() {
		uploaderConfig = uploader.Config{
			FilepathPrefix: filepathPrefix,
			Transport:      fakeTransport,
			SourcesDir:     tempDir,
		}

		uploaderClient = uploader.NewClient(uploaderConfig)
	})

	Describe("UploadFile", func() {
		It("invokes the transport with correct args", func() {
			err := uploaderClient.UploadFile(exactGlob)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTransport.UploadCallCount()).To(Equal(1))

			glob, remoteDir, sourcesDir := fakeTransport.UploadArgsForCall(0)
			Expect(glob).To(Equal(exactGlob))
			Expect(remoteDir).To(Equal(filepathPrefix + "/"))
			Expect(sourcesDir).To(Equal(tempDir))
		})

		Context("when the transport exits with error", func() {
			BeforeEach(func() {
				fakeTransport.UploadReturns(errors.New("some error"))
			})

			It("propagates errors", func() {
				err := uploaderClient.UploadFile("foo")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("some error"))
			})
		})

		Context("when the glob is empty", func() {
			It("returns an error", func() {
				err := uploaderClient.UploadFile("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("glob"))
			})
		})
	})

	Describe("ComputeAWSObjectKey", func() {
		It("computes the correct aws object key", func() {
			remotePath, remoteDir, err := uploaderClient.ComputeAWSObjectKey(exactGlob)

			Expect(err).NotTo(HaveOccurred())
			Expect(remotePath).To(Equal(fmt.Sprint(filepathPrefix, "/", exactGlob)))
			Expect(remoteDir).To(Equal(fmt.Sprint(filepathPrefix, "/")))
		})

		Context("file path Prefix starts with a '/'", func() {
			It("removes the '/' form the prefix", func() {
				filepathPrefix = "/product-files/my-product-slug"
				expectedFilePathPrefix := "product-files/my-product-slug"
				remotePath, remoteDir, err := uploaderClient.ComputeAWSObjectKey(exactGlob)

				Expect(err).NotTo(HaveOccurred())
				Expect(remotePath).To(Equal(fmt.Sprint(expectedFilePathPrefix, "/", exactGlob)))
				Expect(remoteDir).To(Equal(fmt.Sprint(expectedFilePathPrefix, "/")))
			})
		})
	})
})
