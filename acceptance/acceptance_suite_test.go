package acceptance

import (
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/robdimsdale/sanitizer"

	"testing"
)

var (
	inPath    string
	checkPath string
	outPath   string

	endpoint string

	productSlug        string
	pivnetAPIToken     string
	awsAccessKeyID     string
	awsSecretAccessKey string
	pivnetRegion       string
	pivnetBucketName   string
	s3FilepathPrefix   string

	pivnetClient pivnet.Client
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

	By("Getting aws access key id from environment variables")
	awsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	Expect(awsAccessKeyID).NotTo(BeEmpty(), "$AWS_ACCESS_KEY_ID must be provided")

	By("Getting aws secret access key from environment variables")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	Expect(awsSecretAccessKey).NotTo(BeEmpty(), "$AWS_SECRET_ACCESS_KEY must be provided")

	By("Getting pivnet region from environment variables")
	pivnetRegion = os.Getenv("PIVNET_S3_REGION")
	Expect(pivnetRegion).NotTo(BeEmpty(), "$PIVNET_S3_REGION must be provided")

	By("Getting pivnet bucket name from environment variables")
	pivnetBucketName = os.Getenv("PIVNET_BUCKET_NAME")
	Expect(pivnetBucketName).NotTo(BeEmpty(), "$PIVNET_BUCKET_NAME must be provided")

	By("Getting s3 filepath prefix from environment variables")
	s3FilepathPrefix = os.Getenv("S3_FILEPATH_PREFIX")
	Expect(s3FilepathPrefix).NotTo(BeEmpty(), "$S3_FILEPATH_PREFIX must be provided")

	By("Getting endpoint from environment variables")
	endpoint = os.Getenv("PIVNET_ENDPOINT")
	Expect(endpoint).NotTo(BeEmpty(), "$PIVNET_ENDPOINT must be provided")

	By("Compiling check binary")
	checkPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	outPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/out", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling in binary")
	inPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/in", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Copying s3-out to compilation location")
	originalS3OutPath := os.Getenv("S3_OUT_LOCATION")
	Expect(originalS3OutPath).ToNot(BeEmpty(), "$S3_OUT_LOCATION must be provided")
	_, err = os.Stat(originalS3OutPath)
	Expect(err).NotTo(HaveOccurred())
	s3OutPath := filepath.Join(path.Dir(outPath), "s3-out")
	copyFileContents(originalS3OutPath, s3OutPath)
	Expect(err).NotTo(HaveOccurred())

	By("Ensuring copy of s3-out is executable")
	err = os.Chmod(s3OutPath, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	By("Sanitizing acceptance test output")
	sanitized := map[string]string{
		pivnetAPIToken:     "***sanitized-api-token***",
		awsAccessKeyID:     "***sanitized-aws-access-key-id***",
		awsSecretAccessKey: "***sanitized-aws-secret-access-key***",
	}
	sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)
	GinkgoWriter = sanitizer

	By("Creating pivnet client (for out-of-band operations)")
	testLogger := logger.NewLogger(GinkgoWriter)

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     pivnetAPIToken,
		UserAgent: "pivnet-resource/integration-test",
	}
	pivnetClient = pivnet.NewClient(clientConfig, testLogger)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
