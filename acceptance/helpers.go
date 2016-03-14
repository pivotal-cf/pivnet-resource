package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

func getReleases(productSlug string) []pivnet.Release {
	productURL := fmt.Sprintf(
		"%s/api/v2/products/%s/releases",
		endpoint,
		productSlug,
	)

	req, err := http.NewRequest("GET", productURL, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	response := pivnet.ReleasesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())

	return response.Releases
}

func getProductVersions(productSlug string) []string {
	var versions []string
	for _, release := range getReleases(productSlug) {
		versions = append(versions, string(release.Version))
	}

	return versions
}

func getPivnetRelease(productSlug, productVersion string) pivnet.Release {
	for _, release := range getReleases(productSlug) {
		if release.Version == productVersion {
			return release
		}
	}
	Fail(fmt.Sprintf(
		"Could not find release for productSlug: %s and productVersion: %s",
		productSlug,
		productVersion,
	))
	// We won't get here
	return pivnet.Release{}
}

func deletePivnetRelease(productSlug, productVersion string) {
	pivnetRelease := getPivnetRelease(productSlug, productVersion)
	releaseID := pivnetRelease.ID
	Expect(releaseID).NotTo(Equal(0))

	productURL := fmt.Sprintf(
		"%s/api/v2/products/%s/releases/%d",
		endpoint,
		productSlug,
		releaseID,
	)

	req, err := http.NewRequest("DELETE", productURL, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
}

func getProductFiles(productSlug string) []pivnet.ProductFile {
	productURL := fmt.Sprintf(
		"%s/api/v2/products/%s/product_files",
		endpoint,
		productSlug,
	)

	req, err := http.NewRequest("GET", productURL, nil)
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

func getUserGroups(productSlug string, releaseID int) []pivnet.UserGroup {
	userGroupsURL := fmt.Sprintf(
		"%s/api/v2/products/%s/releases/%d/user_groups",
		endpoint,
		productSlug,
		releaseID,
	)

	req, err := http.NewRequest("GET", userGroupsURL, nil)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", pivnetAPIToken))

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	response := pivnet.UserGroups{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())

	return response.UserGroups
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

func run(command *exec.Cmd, stdinContents []byte) *gexec.Session {
	fmt.Fprintf(GinkgoWriter, "input: %s\n", stdinContents)

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

type s3client struct {
	client  *s3.S3
	session *session.Session
}

func NewS3Client(
	accessKey string,
	secretKey string,
	regionName string,
	endpoint string,
) (*s3client, error) {
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")

	awsConfig := &aws.Config{
		Region:           aws.String(regionName),
		Credentials:      creds,
		S3ForcePathStyle: aws.Bool(true),
	}

	sess := session.New(awsConfig)
	client := s3.New(sess, awsConfig)

	return &s3client{
		client:  client,
		session: sess,
	}, nil
}

func (client *s3client) DownloadFile(
	bucketName string,
	remotePath string,
	localPath string,
) error {
	downloader := s3manager.NewDownloader(client.session)

	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	getObject := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(remotePath),
	}

	_, err = downloader.Download(localFile, getObject)
	if err != nil {
		return err
	}

	return nil
}

func (client *s3client) DeleteFile(bucketName string, remotePath string) error {
	_, err := client.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(remotePath),
	})

	return err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
