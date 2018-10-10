package acceptance

import (
	"log"
	"os"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logshim"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/robdimsdale/sanitizer"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	inPath    string
	checkPath string
	outPath   string

	endpoint string

	productSlug        string
	pivnetAPIToken     string
	pivnetRegion       string
	pivnetBucketName   string
	refreshToken       string

	pivnetClient *gp.Client
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var _ = BeforeSuite(func() {
	var err error
	By("Getting product slug from environment variables")
	productSlug = os.Getenv("PRODUCT_SLUG")
	Expect(productSlug).NotTo(BeEmpty(), "$PRODUCT_SLUG must be provided")

	By("Getting API token from environment variables")
	pivnetAPIToken = os.Getenv("API_TOKEN")
	Expect(pivnetAPIToken).NotTo(BeEmpty(), "$API_TOKEN must be provided")

	By("Getting pivnet region from environment variables")
	pivnetRegion = os.Getenv("PIVNET_S3_REGION")
	Expect(pivnetRegion).NotTo(BeEmpty(), "$PIVNET_S3_REGION must be provided")

	By("Getting pivnet bucket name from environment variables")
	pivnetBucketName = os.Getenv("PIVNET_BUCKET_NAME")
	Expect(pivnetBucketName).NotTo(BeEmpty(), "$PIVNET_BUCKET_NAME must be provided")

	By("Getting endpoint from environment variables")
	endpoint = os.Getenv("PIVNET_ENDPOINT")
	Expect(endpoint).NotTo(BeEmpty(), "$PIVNET_ENDPOINT must be provided")

	By("Getting refresh token from environment variables")
	refreshToken = os.Getenv("PIVNET_RESOURCE_REFRESH_TOKEN")
	Expect(refreshToken).NotTo(BeEmpty(), "$PIVNET_RESOURCE_REFRESH_TOKEN must be provided")

	By("Compiling check binary")
	checkPath, err = gexec.Build("github.com/pivotal-cf/pivnet-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	outPath, err = gexec.Build("github.com/pivotal-cf/pivnet-resource/cmd/out", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling in binary")
	inPath, err = gexec.Build("github.com/pivotal-cf/pivnet-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())

	By("Sanitizing acceptance test output")
	sanitized := map[string]string{
		pivnetAPIToken:     "***sanitized-api-token***",
		refreshToken:       "***sanitized-refresh-token***",
	}
	sanitizedWriter := sanitizer.NewSanitizer(sanitized, GinkgoWriter)
	GinkgoWriter = sanitizedWriter

	By("Creating pivnet client (for out-of-band operations)")

	testLogger := log.New(sanitizedWriter, "", log.LstdFlags)
	verbose := true
	ls := logshim.NewLogShim(testLogger, testLogger, verbose)

	clientConfig := pivnet.ClientConfig{
		Host:      endpoint,
		Token:     pivnetAPIToken,
		UserAgent: "pivnet-resource/integration-test",
	}

	pivnetClient = gp.NewClient(clientConfig, ls)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
