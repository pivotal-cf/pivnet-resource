package s3_test

import (
	"github.com/pivotal-cf/pivnet-resource/v3/s3/s3fakes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/v3/s3"
)

var _ = Describe("S3 Client", func() {
	var (
		client *s3.Client
		fakeFileSizeGetter *s3fakes.FakeFileSizeGetter
	)

	BeforeEach(func() {
		fakeFileSizeGetter = &s3fakes.FakeFileSizeGetter{}
		client = s3.NewClient(s3.NewClientConfig{FileSizeGetter: fakeFileSizeGetter})
	})

	Describe("Upload file", func() {
		var (
			sourcesDir string
			fileGlob   string
			to         string
		)

		BeforeEach(func() {
			fileGlob = "some-file*"
			to = "some-remote-file-name"

			var err error
			sourcesDir, err = ioutil.TempDir("", "pivnet-resource-s3-test")
			Expect(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(
				filepath.Join(sourcesDir, fileGlob),
				nil,
				os.ModePerm,
			)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			By("Removing local temp files")
			err := os.RemoveAll(sourcesDir)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when glob is badly-formed", func() {
			BeforeEach(func() {
				fileGlob = "["
			})

			It("returns error", func() {
				err := client.Upload(fileGlob, to, sourcesDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when glob does not match anything", func() {
			BeforeEach(func() {
				fileGlob = "this-will-not-match"
			})

			It("returns error", func() {
				err := client.Upload(fileGlob, to, sourcesDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when glob matches more than one file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(
					filepath.Join(sourcesDir, "some-file-2"),
					nil,
					os.ModePerm,
				)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("returns error", func() {
				err := client.Upload(fileGlob, to, sourcesDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when file size is greater than 20gb", func() {
			BeforeEach(func() {
				fakeFileSizeGetter.FileSizeReturns(25000000000, nil)
			})
			It("returns error", func() {
				err := client.Upload(fileGlob, to, sourcesDir)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
