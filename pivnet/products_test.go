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

var _ = Describe("PivnetClient - product", func() {
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

	Describe("FindProductForSlug", func() {
		var (
			slug = "my-product"
		)
		Context("when the product can be found", func() {
			It("returns the located product", func() {
				response := fmt.Sprintf(`{"id": 3, "slug": "%s"}`, slug)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				product, err := client.FindProductForSlug(slug)
				Expect(err).NotTo(HaveOccurred())
				Expect(product.Slug).To(Equal(slug))
			})
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.FindProductForSlug(slug)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
