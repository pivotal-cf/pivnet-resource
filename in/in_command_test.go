package in_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

var _ = Describe("In", func() {
	var (
		server *ghttp.Server

		downloadDir string

		version      string
		ginkgoLogger logger.Logger

		inRequest concourse.InRequest
		inCommand *in.InCommand
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		var err error
		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		version = "some-version"

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				Endpoint:    server.URL(),
			},
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		inCommand = in.NewInCommand(version, ginkgoLogger, downloadDir)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no api token is provided", func() {
		BeforeEach(func() {
			inRequest.Source.APIToken = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
		})
	})
})
