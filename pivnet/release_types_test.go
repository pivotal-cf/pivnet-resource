package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger/loggerfakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

var _ = Describe("PivnetClient - release types", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.NewClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig = pivnet.NewClientConfig{
			Endpoint:  apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleaseTypes", func() {
		It("returns the release types", func() {
			response := `{"release_types": ["foo","bar"]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			releaseTypes, err := client.ReleaseTypes()
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseTypes).To(HaveLen(2))
			Expect(releaseTypes[0]).To(Equal("foo"))
			Expect(releaseTypes[1]).To(Equal("bar"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.ReleaseTypes()
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
