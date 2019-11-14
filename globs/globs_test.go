package globs_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
	"github.com/pivotal-cf/go-pivnet/v3/logshim"
	"github.com/pivotal-cf/pivnet-resource/globs"
)

var _ = Describe("Globber", func() {
	Describe("ExactGlobs", func() {
		var (
			fakeLogger logger.Logger

			globberConfig globs.GlobberConfig
			globber       *globs.Globber

			tempDir    string
			myFilesDir string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "pivnet-resource")
			Expect(err).NotTo(HaveOccurred())

			myFilesDir = filepath.Join(tempDir, "my_files")
			err = os.Mkdir(myFilesDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Create(filepath.Join(myFilesDir, "file-0"))
			Expect(err).NotTo(HaveOccurred())

			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			globberConfig = globs.GlobberConfig{
				FileGlob:   "my_files/*",
				SourcesDir: tempDir,
				Logger:     fakeLogger,
			}

			globber = globs.NewGlobber(globberConfig)
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no files match the fileglob", func() {
			BeforeEach(func() {
				globberConfig.FileGlob = "this-will-match-nothing"
				globber = globs.NewGlobber(globberConfig)
			})

			It("returns an error", func() {
				_, err := globber.ExactGlobs()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("no matches"))
			})
		})

		Context("when multiple files match the fileglob", func() {
			BeforeEach(func() {
				_, err := os.Create(filepath.Join(myFilesDir, "file-1"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a map of filenames to remote paths", func() {
				filenamePaths, err := globber.ExactGlobs()
				Expect(err).NotTo(HaveOccurred())

				Expect(len(filenamePaths)).To(Equal(2))

				Expect(filenamePaths[0]).To(Equal("my_files/file-0"))
				Expect(filenamePaths[1]).To(Equal("my_files/file-1"))
			})
		})
	})
})
