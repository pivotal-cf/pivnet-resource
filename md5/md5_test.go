package md5_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/md5"
)

var _ = Describe("MD5", func() {
	Describe("FileContentsSummer", func() {
		var (
			tempFilePath string
			tempDir      string
			fileContents []byte

			summer md5.Summer
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			fileContents = []byte("foobar contents")

			tempFilePath = filepath.Join(tempDir, "foobar")

			ioutil.WriteFile(tempFilePath, fileContents, os.ModePerm)

			summer = md5.NewFileContentsSummer(tempFilePath)
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the MD5 of a file without error", func() {
			md5, err := summer.Sum()
			Expect(err).NotTo(HaveOccurred())

			// Expected md5 of 'foobar contents'
			Expect(md5).To(Equal("fdd3d599138fd15d7673f3d3539531c1"))
		})

		Context("when there is an error reading the file", func() {
			BeforeEach(func() {
				tempFilePath = "/not/a/valid/file"

				summer = md5.NewFileContentsSummer(tempFilePath)
			})

			It("returns the error", func() {
				_, err := summer.Sum()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
