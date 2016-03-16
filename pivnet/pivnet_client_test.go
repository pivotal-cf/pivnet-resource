package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

var _ = Describe("PivnetClient", func() {
	var (
		server    *ghttp.Server
		client    pivnet.Client
		token     string
		userAgent string

		releases   pivnet.ReleasesResponse
		etagHeader []http.Header

		newClientConfig pivnet.NewClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		releases = pivnet.ReleasesResponse{Releases: []pivnet.Release{
			{
				ID:      1,
				Version: "1234",
			},
			{
				ID:      99,
				Version: "some-other-version",
			},
		}}

		etagHeader = []http.Header{
			{"ETag": []string{`"etag-0"`}},
			{"ETag": []string{`"etag-1"`}},
		}

		server = ghttp.NewServer()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = logger.NewLogger(GinkgoWriter)
		// fakeLogger = &logger_fakes.FakeLogger{}
		newClientConfig = pivnet.NewClientConfig{
			Endpoint:  server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	It("has authenticated headers for each request", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for i, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.RespondWith(http.StatusOK, nil, etagHeader[i]),
				),
			)
		}

		_, err := client.ProductVersions(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets custom user agent", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.VerifyHeaderKV("User-Agent", userAgent),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for i, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyHeaderKV("User-Agent", userAgent),
					ghttp.RespondWith(http.StatusOK, nil, etagHeader[i]),
				),
			)
		}

		_, err := client.ProductVersions(productSlug)
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
				newClientConfig.Endpoint = "%%%"
				client = pivnet.NewClient(newClientConfig, fakeLogger)

				_, err := client.ProductVersions("some product")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("%%%"))
			})
		})

		Context("when making the request fails with error", func() {
			It("forwards the error", func() {
				newClientConfig.Endpoint = "https://not-a-real-url.com"
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
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
				),
			)

			for i, r := range releases.Releases {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
						),
						ghttp.RespondWith(http.StatusOK, nil, etagHeader[i]),
					),
				)
			}

			// one for all the releases and one for each releases
			expectedRequestCount := 1 + 2

			versions, err := client.ProductVersions(productSlug)
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(expectedRequestCount))
			Expect(versions).To(HaveLen(len(releases.Releases)))
			Expect(versions[0]).Should(Equal(releases.Releases[0].Version + "#etag-0"))
			Expect(versions[1]).Should(Equal(releases.Releases[1].Version + "#etag-1"))
		})
	})
})
