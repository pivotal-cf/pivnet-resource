package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	"testing"
)

var (
	inPath    string
	checkPath string
	outPath   string

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

	By("Compiling check binary")
	checkPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	outPath, err = gexec.Build("github.com/pivotal-cf-experimental/pivnet-resource/cmd/out", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	inPath, err = gexec.Build(
		"github.com/pivotal-cf-experimental/pivnet-resource/cmd/in", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Copying s3-out to compilation location")
	originalS3OutPath := os.Getenv("S3_OUT_LOCATION")
	Expect(originalS3OutPath).ToNot(BeEmpty(), "$S3_OUT_LOCATION must be provided")
	s3OutPath := filepath.Join(path.Dir(outPath), "s3-out")
	copyFileContents(originalS3OutPath, s3OutPath)
	Expect(err).NotTo(HaveOccurred())

	By("Ensuring copy of s3-out is executable")
	err = os.Chmod(s3OutPath, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	By("Creating pivnet client (for out-of-band operations)")
	testLogger := logger.NewLogger(GinkgoWriter)

	clientConfig := pivnet.NewClientConfig{
		URL:       pivnet.URL,
		Token:     pivnetAPIToken,
		UserAgent: "pivnet-resource/integration-test",
	}
	pivnetClient = pivnet.NewClient(clientConfig, testLogger)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func getProductReleases(productSlug string) []pivnet.Release {
	productURL := fmt.Sprintf(
		"https://network.pivotal.io/api/v2/products/%s/releases",
		productSlug)

	req, err := http.NewRequest("GET", productURL, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	response := pivnet.Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())

	return response.Releases
}

func getProductVersions(productSlug string) []string {
	var versions []string
	for _, release := range getProductReleases(productSlug) {
		versions = append(versions, string(release.Version))
	}

	return versions
}

func getPivnetRelease(productSlug, productVersion string) pivnet.Release {
	for _, release := range getProductReleases(productSlug) {
		if release.Version == productVersion {
			return release
		}
	}
	Fail(fmt.Sprintf("Could not find release for productSlug: %s and productVersion: %s", productSlug, productVersion))
	// We won't get here
	return pivnet.Release{}
}

func deletePivnetRelease(productSlug, productVersion string) {
	pivnetRelease := getPivnetRelease(productSlug, productVersion)
	releaseID := pivnetRelease.ID
	Expect(releaseID).NotTo(Equal(0))

	product_url := fmt.Sprintf("https://network.pivotal.io/api/v2/products/%s/releases/%d", productSlug, releaseID)

	req, err := http.NewRequest("DELETE", product_url, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
}

func getProductFiles(productSlug string) []pivnet.ProductFile {
	product_url := fmt.Sprintf("https://network.pivotal.io/api/v2/products/%s/product_files", productSlug)

	req, err := http.NewRequest("GET", product_url, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	response := pivnet.ProductFiles{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())

	return response.ProductFiles
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// See http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func sanitize(contents string) string {
	output := contents
	output = strings.Replace(output, pivnetAPIToken, "***sanitized-api-token***", -1)
	output = strings.Replace(output, awsAccessKeyID, "***sanitized-aws-access-key-id***", -1)
	output = strings.Replace(output, awsSecretAccessKey, "***sanitized-aws-secret-access-key***", -1)
	return output
}

func run(command *exec.Cmd, stdinContents []byte) *gexec.Session {
	fmt.Fprintf(GinkgoWriter, "input: %s\n", sanitize(string(stdinContents)))

	stdin, err := command.StdinPipe()
	Expect(err).ShouldNot(HaveOccurred())

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	_, err = io.WriteString(stdin, string(stdinContents))
	Expect(err).ShouldNot(HaveOccurred())

	err = stdin.Close()
	Expect(err).ShouldNot(HaveOccurred())

	return session
}

func metadataValueForKey(metadata []concourse.Metadata, name string) (string, error) {
	for _, i := range metadata {
		if i.Name == name {
			return i.Value, nil
		}
	}
	return "", fmt.Errorf("name not found: %s", name)
}
