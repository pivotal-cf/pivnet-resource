package sanitizer_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

var _ = Describe("Sanitizer", func() {
	var (
		tempDir     string
		logFilepath string
		logFile     *os.File

		pairs map[string]string
		s     sanitizer.Sanitizer
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		logFilepath = filepath.Join(tempDir, "debug.log")
		logFile, err = os.Create(logFilepath)
		Expect(err).NotTo(HaveOccurred())

		pairs = make(map[string]string)

		s = sanitizer.NewSanitizer(pairs, logFile)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Write", func() {
		It("sanitizes correctly", func() {
			pairs["secret_value"] = "***secret-redacted***"
			pairs["super_secret_value"] = "***super-secret-redacted***"
			_, err := s.Write([]byte("my secret is: secret_value"))
			Expect(err).NotTo(HaveOccurred())

			err = logFile.Sync()
			Expect(err).NotTo(HaveOccurred())

			err = logFile.Close()
			Expect(err).NotTo(HaveOccurred())

			b, err := ioutil.ReadFile(logFilepath)
			Expect(err).NotTo(HaveOccurred())
			Expect(b).To(Equal([]byte("my secret is: ***secret-redacted***")))
		})
	})
})
