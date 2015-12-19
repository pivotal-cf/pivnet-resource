package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

const (
	s3UploadTimeout   = 5 * time.Second
	executableTimeout = 5 * time.Second
)

type s3client struct {
	client  *s3.S3
	session *session.Session
}

func run(command *exec.Cmd, stdinContents []byte) *gexec.Session {
	fmt.Fprintln(GinkgoWriter, sanitize(string(stdinContents)))

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

var _ = Describe("Out", func() {
	var (
		versionFile *os.File

		releaseTypeFile *os.File
		releaseType     = "Minor Release"

		releaseDateFile *os.File
		releaseDate     = "2015-12-17"

		eulaSlugFile *os.File
		eulaSlug     = "pivotal_beta_eula"

		productVersion string
		productName    = "pivotal-diego-pcf"

		command       *exec.Cmd
		stdinContents []byte
		outRequest    concourse.OutRequest
		sourcesDir    string
	)

	BeforeEach(func() {
		var err error
		By("Generating 'random' product version")
		productVersion = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Writing product version to file")
		versionFile, err = ioutil.TempFile("", "")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = versionFile.WriteString(productVersion)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing release type to file")
		releaseTypeFile, err = ioutil.TempFile("", "")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = releaseTypeFile.WriteString(releaseType)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing release date to file")
		releaseDateFile, err = ioutil.TempFile("", "")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = releaseDateFile.WriteString(releaseDate)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing eula slug to file")
		eulaSlugFile, err = ioutil.TempFile("", "")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = eulaSlugFile.WriteString(eulaSlug)
		Expect(err).ShouldNot(HaveOccurred())

		By("Creating a temporary sources dir")
		sourcesDir, err = ioutil.TempDir("", "")
		Expect(err).ShouldNot(HaveOccurred())

		command = exec.Command(outPath, sourcesDir)

		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        pivnetAPIToken,
				AccessKeyID:     awsAccessKeyID,
				SecretAccessKey: awsSecretAccessKey,
				ProductName:     productName,
			},
			Params: concourse.OutParams{
				File:            "*",
				FilepathPrefix:  s3FilepathPrefix,
				VersionFile:     versionFile.Name(),
				ReleaseTypeFile: releaseTypeFile.Name(),
				ReleaseDateFile: releaseDateFile.Name(),
				EulaSlugFile:    eulaSlugFile.Name(),
			},
		}

		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing local temp version file")
		err := os.RemoveAll(versionFile.Name())
		Expect(err).ShouldNot(HaveOccurred())

		By("Removing local temp release type file")
		err = os.RemoveAll(releaseTypeFile.Name())
		Expect(err).ShouldNot(HaveOccurred())

		By("Removing local temp release date file")
		err = os.RemoveAll(releaseDateFile.Name())
		Expect(err).ShouldNot(HaveOccurred())

		By("Removing local temp eula slug file")
		err = os.RemoveAll(eulaSlugFile.Name())
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Argument validation", func() {
		Context("when no sources directory is provided via args", func() {
			It("exits with error", func() {
				command := exec.Command(outPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("sources"))
			})
		})

		Context("when no api_token is provided", func() {
			BeforeEach(func() {
				outRequest.Source.APIToken = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("api_token must be provided"))
			})
		})

		Context("when no product_name is provided", func() {
			BeforeEach(func() {
				outRequest.Source.ProductName = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("product_name must be provided"))
			})
		})

		Context("when no aws access key id is provided", func() {
			BeforeEach(func() {
				outRequest.Source.AccessKeyID = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("access_key_id must be provided"))
			})
		})

		Context("when no aws secret access key is provided", func() {
			BeforeEach(func() {
				outRequest.Source.SecretAccessKey = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("secret_access_key must be provided"))
			})
		})

		Context("when no file glob is provided", func() {
			BeforeEach(func() {
				outRequest.Params.File = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("file glob must be provided"))
			})
		})

		Context("when no s3 filepath prefix is provided", func() {
			BeforeEach(func() {
				outRequest.Params.FilepathPrefix = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("s3_filepath_prefix must be provided"))
			})
		})

		Context("when no version_file is provided", func() {
			BeforeEach(func() {
				outRequest.Params.VersionFile = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("version_file must be provided"))
			})
		})

		Context("when no release_type_file is provided", func() {
			BeforeEach(func() {
				outRequest.Params.ReleaseTypeFile = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("release_type_file must be provided"))
			})
		})

		Context("when no eula_slug_file is provided", func() {
			BeforeEach(func() {
				outRequest.Params.EulaSlugFile = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("eula_slug_file must be provided"))
			})
		})
	})

	Describe("Creating a new release", func() {
		AfterEach(func() {
			By("Deleting newly-created release")
			deletePivnetRelease(productName, productVersion)
		})

		It("Successfully creates a release", func() {
			outRequest.Params.File = ""
			outRequest.Params.FilepathPrefix = ""

			var err error
			stdinContents, err = json.Marshal(outRequest)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the new product version does not yet exist")
			productVersions := getProductVersions(productName)
			Expect(productVersions).NotTo(BeEmpty())
			Expect(productVersions).NotTo(ContainElement(productVersion))

			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			By("Validating new release exists on pivnet")
			productVersions = getProductVersions("pivotal-diego-pcf")
			Expect(productVersions).To(ContainElement(productVersion))

			By("Outputting a valid json response")
			response := concourse.OutResponse{}
			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response.Version.ProductVersion).To(Equal(productVersion))

			By("Validating the release was created correctly")
			release := getPivnetRelease(productName, productVersion)
			Expect(release.Version).To(Equal(productVersion))
			Expect(release.ReleaseType).To(Equal(releaseType))
			Expect(release.ReleaseDate).To(Equal(releaseDate))
			Expect(release.Eula.Slug).To(Equal(eulaSlug))
		})

		Context("when no release_date_file is provided", func() {
			BeforeEach(func() {
				outRequest.Params.ReleaseDateFile = ""

				outRequest.Params.File = ""
				outRequest.Params.FilepathPrefix = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("Defaults the release_date to the current date", func() {
				todayDate := time.Now().Format("2006-01-02")

				session := run(command, stdinContents)

				Eventually(session, executableTimeout).Should(gexec.Exit(0))

				release := getPivnetRelease(productName, productVersion)
				Expect(release.ReleaseDate).To(Equal(todayDate))
			})
		})

		Context("when S3 source and params are configured correctly", func() {
			var (
				client *s3client

				sourceFileName string
				sourceFilePath string
				remotePath     string
			)

			BeforeEach(func() {
				By("Creating aws client")
				var err error
				client, err = NewS3Client(
					awsAccessKeyID,
					awsSecretAccessKey,
					pivnetRegion,
					pivnetBucketName,
				)
				Expect(err).ShouldNot(HaveOccurred())

				sourceFileName = fmt.Sprintf("pivnet-resource-test-file-%d", time.Now().Nanosecond())

				By("Creating local temp files")
				sourceFilePath = filepath.Join(sourcesDir, sourceFileName)
				err = ioutil.WriteFile(sourceFilePath, []byte("some content"), os.ModePerm)
				Expect(err).ShouldNot(HaveOccurred())

				remotePath = fmt.Sprintf("product_files/%s/%s", s3FilepathPrefix, sourceFileName)
			})

			AfterEach(func() {
				By("Removing uploaded file")
				client.DeleteFile(pivnetBucketName, remotePath)

				By("Removing local temp files")
				err := os.RemoveAll(sourcesDir)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("uploads a single file to s3", func() {
				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, s3UploadTimeout).Should(gexec.Exit(0))

				By("Verifying uploaded file can be downloaded")
				localDownloadPath := fmt.Sprintf("%s-downloaded", sourceFilePath)
				err := client.DownloadFile(pivnetBucketName, remotePath, localDownloadPath)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})

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
