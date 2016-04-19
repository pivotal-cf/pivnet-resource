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

var _ = Describe("PivnetClient - release dependencies", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.NewClientConfig
		fakeLogger      logger.Logger

		productID int
		releaseID int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		productID = 1234
		releaseID = 2345

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

	Describe("ReleaseDependencies", func() {
		It("returns the release dependencies", func() {

			response := pivnet.ReleaseDependenciesResponse{
				ReleaseDependencies: []pivnet.ReleaseDependency{
					{
						Release: pivnet.DependentRelease{
							ID:      9876,
							Version: "release 9876",
							Product: pivnet.Product{
								ID:   23,
								Name: "Product 23",
							},
						},
					},
					{
						Release: pivnet.DependentRelease{
							ID:      8765,
							Version: "release 8765",
							Product: pivnet.Product{
								ID:   23,
								Name: "Product 23",
							},
						},
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%d/releases/%d/dependencies",
						apiPrefix,
						productID,
						releaseID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			releaseDependencies, err := client.ReleaseDependencies(productID, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseDependencies).To(HaveLen(2))
			Expect(releaseDependencies[0].Release.ID).To(Equal(9876))
			Expect(releaseDependencies[1].Release.ID).To(Equal(8765))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%d/releases/%d/dependencies",
							apiPrefix,
							productID,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.ReleaseDependencies(productID, releaseID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
