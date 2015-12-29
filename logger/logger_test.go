package logger_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

var _ = Describe("Logger", func() {
	var (
		tempDir     string
		logFilepath string
		logFile     *os.File
		l           logger.Logger
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		logFilepath = filepath.Join(tempDir, "debug.log")
		logFile, err = os.Create(logFilepath)
		Expect(err).NotTo(HaveOccurred())

		l = logger.NewLogger(logFile)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Debugf", func() {
		Context("when sink is writable", func() {
			It("logs correctly", func() {
				currentTime := time.Now().Nanosecond()
				_, err := l.Debugf("current time: %d\n", currentTime)
				Expect(err).NotTo(HaveOccurred())

				b, err := ioutil.ReadFile(logFilepath)
				Expect(err).NotTo(HaveOccurred())
				Expect(b).To(Equal([]byte(fmt.Sprintf("current time: %d\n", currentTime))))
			})
		})
		Context("when sink is not writable", func() {
			BeforeEach(func() {
				err := logFile.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns error", func() {
				_, err := l.Debugf("some contents")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
