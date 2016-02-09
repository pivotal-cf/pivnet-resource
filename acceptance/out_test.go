package acceptance

import (
	"encoding/json"
	"fmt"
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
	executableTimeout = 60 * time.Second
)

type s3client struct {
	client  *s3.S3
	session *session.Session
}

var _ = Describe("Out", func() {
	var (
		releaseTypeFile = "release_type"
		releaseType     = "Minor Release"

		releaseDateFile = "release_date"
		releaseDate     = "2015-12-17"

		eulaSlugFile = "eula_slug"
		eulaSlug     = "pivotal_beta_eula"

		productVersionFile = "version"
		productVersion     string

		descriptionFile = "description"
		description     = "this release is for automated-testing only."

		releaseNotesURLFile = "release_notes_url"
		releaseNotesURL     = "https://example.com"

		command       *exec.Cmd
		stdinContents []byte
		outRequest    concourse.OutRequest
		rootDir       string
	)

	BeforeEach(func() {
		var err error

		By("Creating a temporary root dir")
		rootDir, err = ioutil.TempDir("", "")
		Expect(err).ShouldNot(HaveOccurred())

		By("Generating 'random' product version")
		productVersion = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Writing product version to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, productVersionFile),
			[]byte(productVersion),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing release type to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, releaseTypeFile),
			[]byte(releaseType),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing release date to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, releaseDateFile),
			[]byte(releaseDate),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing eula slug to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, eulaSlugFile),
			[]byte(eulaSlug),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing description to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, descriptionFile),
			[]byte(description),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing release notes URL to file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, releaseNotesURLFile),
			[]byte(releaseNotesURL),
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Creating command object")
		command = exec.Command(outPath, rootDir)

		By("Creating default request")
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        pivnetAPIToken,
				AccessKeyID:     awsAccessKeyID,
				SecretAccessKey: awsSecretAccessKey,
				ProductSlug:     productSlug,
				Endpoint:        endpoint,
				Bucket:          pivnetBucketName,
				Region:          pivnetRegion,
			},
			Params: concourse.OutParams{
				FileGlob:            "",
				FilepathPrefix:      "",
				VersionFile:         productVersionFile,
				ReleaseTypeFile:     releaseTypeFile,
				ReleaseDateFile:     releaseDateFile,
				EulaSlugFile:        eulaSlugFile,
				DescriptionFile:     descriptionFile,
				ReleaseNotesURLFile: releaseNotesURLFile,
			},
		}

		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing local temp files")
		err := os.RemoveAll(rootDir)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Argument validation", func() {
		Context("when no root directory is provided via args", func() {
			It("exits with error", func() {
				command := exec.Command(outPath)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("usage"))
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

		Context("when no product_slug is provided", func() {
			BeforeEach(func() {
				outRequest.Source.ProductSlug = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("product_slug must be provided"))
			})
		})

		Context("when S3 filepath prefix and glob are provided", func() {
			BeforeEach(func() {
				outRequest.Params.FilepathPrefix = "foo"
				outRequest.Params.FileGlob = "*"

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
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
		})

		Context("when s3 filepath prefix is provided but not file glob", func() {
			BeforeEach(func() {
				outRequest.Params.FilepathPrefix = "foo"

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

		Context("when file glob is provided but not s3 filepath prefix", func() {
			BeforeEach(func() {
				outRequest.Params.FileGlob = "*"

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
			deletePivnetRelease(productSlug, productVersion)
		})

		It("Successfully creates a release", func() {
			var err error
			stdinContents, err = json.Marshal(outRequest)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the new product version does not yet exist")
			productVersions := getProductVersions(productSlug)
			Expect(productVersions).NotTo(ContainElement(productVersion))

			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			By("Validating new release exists on pivnet")
			productVersions = getProductVersions(productSlug)
			Expect(productVersions).To(ContainElement(productVersion))

			By("Outputting a valid json response")
			response := concourse.OutResponse{}
			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response.Version.ProductVersion).To(Equal(productVersion))

			By("Validating the release was created correctly")
			release := getPivnetRelease(productSlug, productVersion)
			Expect(release.Version).To(Equal(productVersion))
			Expect(release.ReleaseType).To(Equal(releaseType))
			Expect(release.ReleaseDate).To(Equal(releaseDate))
			Expect(release.Eula.Slug).To(Equal(eulaSlug))
			Expect(release.Description).To(Equal(description))
			Expect(release.ReleaseNotesURL).To(Equal(releaseNotesURL))

			By("Validing the returned metadata")
			metadataReleaseType, err := metadataValueForKey(response.Metadata, "release_type")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadataReleaseType).To(Equal(releaseType))

			metadataReleaseDate, err := metadataValueForKey(response.Metadata, "release_date")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadataReleaseDate).To(Equal(releaseDate))

			metadataDescription, err := metadataValueForKey(response.Metadata, "description")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadataDescription).To(Equal(description))

			metadataReleaseNotesURL, err := metadataValueForKey(response.Metadata, "release_notes_url")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadataReleaseNotesURL).To(Equal(releaseNotesURL))
		})

		Context("When the availability is set to Selected User Groups Only", func() {
			var (
				availabilityFile = "availability"
				availability     = "Selected User Groups Only"

				userGroupIDsFile = "user_group_ids"
				userGroupIDs     = "6,8,54"
			)

			BeforeEach(func() {
				By("Writing availability to file")
				err := ioutil.WriteFile(
					filepath.Join(rootDir, availabilityFile),
					[]byte(availability),
					os.ModePerm)
				Expect(err).ShouldNot(HaveOccurred())

				By("Writing user group IDs to file")
				err = ioutil.WriteFile(
					filepath.Join(rootDir, userGroupIDsFile),
					[]byte(userGroupIDs),
					os.ModePerm)
				Expect(err).ShouldNot(HaveOccurred())

				outRequest.Params.AvailabilityFile = availabilityFile
				outRequest.Params.UserGroupIDsFile = userGroupIDsFile
			})

			It("Creates a release and updates the availability and user groups", func() {
				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating the new product version does not yet exist")
				productVersions := getProductVersions(productSlug)
				Expect(productVersions).NotTo(ContainElement(productVersion))

				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, executableTimeout).Should(gexec.Exit(0))

				By("Validating new release exists on pivnet")
				productVersions = getProductVersions(productSlug)
				Expect(productVersions).To(ContainElement(productVersion))

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err = json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(response.Version.ProductVersion).To(Equal(productVersion))

				By("Validating the release was created correctly")
				release := getPivnetRelease(productSlug, productVersion)
				Expect(release.Availability).To(Equal(availability))

				By("Validing the returned metadata")
				metadataAvailability, err := metadataValueForKey(response.Metadata, "availability")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(metadataAvailability).To(Equal(availability))

				By("Validating the user groups were associated with the release")
				userGroups := getUserGroups(productSlug, release.ID)
				userGroupIDs := []int{}
				for _, userGroup := range userGroups {
					userGroupIDs = append(userGroupIDs, userGroup.ID)
				}
				Expect(userGroupIDs).Should(ConsistOf(6, 8, 54))
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
