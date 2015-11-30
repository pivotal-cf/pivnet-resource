package pivnet_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/pivotal-cf-experimental/pivnet-resource"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const (
	apiPrefix = "/api/v2"
)

var _ = Describe("PivnetClient", func() {
	var (
		server *ghttp.Server
		client pivnet.Client
		token  string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress := server.URL() + apiPrefix
		token = "my-auth-token"
		client = pivnet.NewClient(apiAddress, token)
	})

	AfterEach(func() {
		server.Close()
	})

	It("has authenticated headers for each request", func() {
		response := fmt.Sprintf(`{"releases": [{"version": "1234"}]}`)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token: %s", token)),
				ghttp.RespondWith(http.StatusOK, response),
			),
		)

		_, err := client.ProductVersions("my-product-id")
		Expect(err).NotTo(HaveOccurred())
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
