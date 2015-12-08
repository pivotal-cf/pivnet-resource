package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var checkPath string

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	Version string `json:"version"`
}

type concourseRequest struct {
	Source  Source   `json:"source"`
	Version struct{} `json:"version"`
}

type Source struct {
	APIToken     string `json:"api_token"`
	ResourceName string `json:"resource_name"`
}

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var _ = BeforeSuite(func() {
	var err error
	By("Compiling binary")
	checkPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func getProductRelease(product string) Release {
	product_url := fmt.Sprintf("https://network.pivotal.io/api/v2/products/%s/releases", product)

	req, err := http.NewRequest("GET", product_url, nil)
	Expect(err).NotTo(HaveOccurred())

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())

	response := Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())

	return response.Releases[0]
}
