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
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress := server.URL() + apiPrefix
		client = pivnet.NewClient(apiAddress)
	})

	AfterEach(func() {
		server.Close()
	})

	XIt("has authenticated headers for each request", func() {
	})

	It("gets versions", func() {
		proudctVersion := "v" + strconv.Itoa(rand.Int())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id"),
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{"products": [{"id": %s}]}`, productVersion)),
			),
		)
		versions, err := client.ProductVersions("my-product-id")
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).Should(HaveLen(1))
		Expect(versions[0]).Should(Equal(productVersion))
	})
})
