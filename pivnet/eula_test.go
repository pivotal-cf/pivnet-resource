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

var _ = Describe("PivnetClient - EULA", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.NewClientConfig
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"
		logger := lager.NewLogger("doesn't matter")

		newClientConfig = pivnet.NewClientConfig{
			Endpoint:  apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}

		client = pivnet.NewClient(newClientConfig, logger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("EULAs", func() {
		It("returns the EULAs", func() {
			response := `{"eulas": [{"id":1,"name":"eula1"},{"id": 2,"name":"eula2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			eulas, err := client.EULAs()
			Expect(err).NotTo(HaveOccurred())

			Expect(eulas).To(HaveLen(2))

			Expect(eulas[0].ID).To(Equal(1))
			Expect(eulas[0].Name).To(Equal("eula1"))
			Expect(eulas[1].ID).To(Equal(2))
			Expect(eulas[1].Name).To(Equal("eula2"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.EULAs()
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("AcceptEULA", func() {
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
})
