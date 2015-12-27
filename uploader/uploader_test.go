package uploader_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
	uploader_fakes "github.com/pivotal-cf-experimental/pivnet-resource/uploader/fakes"
)

var _ = Describe("Uploader", func() {
	var (
		fakeTransport  *uploader_fakes.FakeTransport
		uploaderConfig uploader.Config
		uploaderClient uploader.Client
	)

	BeforeEach(func() {
		fakeTransport = &uploader_fakes.FakeTransport{}

		uploaderConfig = uploader.Config{
			Transport: fakeTransport,
		}

		uploaderClient = uploader.NewClient(uploaderConfig)
	})

	AfterEach(func() {
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
