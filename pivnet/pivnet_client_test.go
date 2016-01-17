package pivnet_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	logger_fakes "github.com/pivotal-cf-experimental/pivnet-resource/logger/fakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

var _ = Describe("PivnetClient", func() {
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
		apiAddress = server.URL() + apiPrefix
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &logger_fakes.FakeLogger{}
		newClientConfig = pivnet.NewClientConfig{
			URL:       apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	It("has authenticated headers for each request", func() {
		response := fmt.Sprintf(`{"releases": [{"version": "1234"}]}`)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.RespondWith(http.StatusOK, response),
			),
		)

		_, err := client.ProductVersions("my-product-id")
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets custom user agent", func() {
		response := fmt.Sprintf(`{"releases": [{"version": "1234"}]}`)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.VerifyHeaderKV("User-Agent", userAgent),
				ghttp.RespondWith(http.StatusOK, response),
			),
		)

		_, err := client.ProductVersions("my-product-id")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Accepting a EULA", func() {
		var (
			releaseID         int
			productSlug       string
			EULAAcceptanceURL string
		)

		BeforeEach(func() {
			productSlug = "banana-slug"
			releaseID = 42
			EULAAcceptanceURL = fmt.Sprintf(apiPrefix+"/products/%s/releases/%d/eula_acceptance", productSlug, releaseID)
		})

		It("accepts the EULA for a given release and product ID", func() {
			response := fmt.Sprintf(`{"accepted_at": "2016-01-11","_links":{}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", EULAAcceptanceURL),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyJSON(`{}`),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			Expect(client.AcceptEULA(productSlug, releaseID)).To(Succeed())
		})

		Context("when any other non-200 status code comes back", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", EULAAcceptanceURL),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
						ghttp.VerifyJSON(`{}`),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				Expect(client.AcceptEULA(productSlug, releaseID)).To(MatchError("Pivnet returned status code: 418 for the request - expected 200"))
			})
		})
	})

	Describe("Product Versions", func() {
		Context("when parsing the url fails with error", func() {
			It("forwards the error", func() {
				newClientConfig.URL = "%%%"
				client = pivnet.NewClient(newClientConfig, fakeLogger)

				_, err := client.ProductVersions("some product")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("%%%"))
			})
		})

		Context("when making the request fails with error", func() {
			It("forwards the error", func() {
				newClientConfig.URL = "https://not-a-real-url.com"
				client = pivnet.NewClient(newClientConfig, fakeLogger)

				_, err := client.ProductVersions("some-product")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a non-200 comes back from Pivnet", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
				)

				_, err := client.ProductVersions("my-product-id")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(
					"Pivnet returned status code: 404 for the request - expected 200"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
						ghttp.RespondWith(http.StatusOK, "%%%"),
					),
				)

				_, err := client.ProductVersions("my-product-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})

		It("gets versions", func() {
			productVersion := "v" + strconv.Itoa(rand.Int())
			response := fmt.Sprintf(
				`{"releases": [{"version": %q}, {"version": %q}]}`,
				productVersion, productVersion,
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			versions, err := client.ProductVersions("my-product-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(1))
			Expect(versions).To(HaveLen(2))
			Expect(versions[0]).Should(Equal(productVersion))
			Expect(versions[1]).Should(Equal(productVersion))
		})
	})
})
