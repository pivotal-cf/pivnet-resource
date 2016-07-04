package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-golang/lager"
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
		logger := lager.NewLogger("doesn't matter")

		newClientConfig = pivnet.NewClientConfig{
			Endpoint:  server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}

		client = pivnet.NewClient(newClientConfig, logger)
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

		_, err := client.ReleasesForProductSlug(productSlug)
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

		_, err := client.ReleasesForProductSlug(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when parsing the url fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Endpoint = "%%%"
			logger := lager.NewLogger("doesn't matter")
			client = pivnet.NewClient(newClientConfig, logger)

			_, err := client.ReleasesForProductSlug("some product")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("%%%"))
		})
	})

	Context("when making the request fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Endpoint = "https://not-a-real-url.com"
			logger := lager.NewLogger("doesn't matter")
			client = pivnet.NewClient(newClientConfig, logger)

			_, err := client.ReleasesForProductSlug("some-product")
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

			_, err := client.ReleasesForProductSlug("my-product-id")
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

			_, err := client.ReleasesForProductSlug("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})

	Describe("Product Versions", func() {
		Context("when getting the ETag responds with a non-2XX status code", func() {
			It("returns an error", func() {
				// server.AppendHandlers(
				// 	ghttp.CombineHandlers(
				// 		ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
				// 		ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
				// 	),
				// )

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, releases.Releases[0].ID)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.ProductVersions("banana", releases.Releases)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})

		It("gets versions", func() {
			// server.AppendHandlers(
			// 	ghttp.CombineHandlers(
			// 		ghttp.VerifyRequest(
			// 			"GET",
			// 			fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
			// 		),
			// 		ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			// 	),
			// )

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

			expectedRequestCount := len(releases.Releases)

			versions, err := client.ProductVersions(productSlug, releases.Releases)
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(expectedRequestCount))
			Expect(versions).To(HaveLen(len(releases.Releases)))
			Expect(versions[0]).Should(Equal(releases.Releases[0].Version + "#etag-0"))
			Expect(versions[1]).Should(Equal(releases.Releases[1].Version + "#etag-1"))
		})
	})
})
