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
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
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
				FileGlob:            "*",
				FilepathPrefix:      s3FilepathPrefix,
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
				outRequest.Params.FileGlob = ""

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
			deletePivnetRelease(productSlug, productVersion)
		})

		It("Successfully creates a release", func() {
			outRequest.Params.FileGlob = ""
			outRequest.Params.FilepathPrefix = ""

			var err error
			stdinContents, err = json.Marshal(outRequest)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the new product version does not yet exist")
			productVersions := getProductVersions(productSlug)
			Expect(productVersions).NotTo(BeEmpty())
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

		Context("when S3 source and params are configured correctly", func() {
			var (
				client *s3client

				sourcesDir      = "sources"
				sourceFileNames []string
				sourceFilePaths []string
				remotePaths     []string

				totalFiles = 3
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

				By("Creating a temporary sources dir")
				sourcesFullPath := filepath.Join(rootDir, sourcesDir)
				err = os.Mkdir(sourcesFullPath, os.ModePerm)
				Expect(err).ShouldNot(HaveOccurred())

				By("Creating local temp files")
				sourceFileNames = make([]string, totalFiles)
				sourceFilePaths = make([]string, totalFiles)
				remotePaths = make([]string, totalFiles)
				for i := 0; i < totalFiles; i++ {
					sourceFileNames[i] = fmt.Sprintf(
						"pivnet-resource-test-file-%d",
						time.Now().Nanosecond(),
					)

					sourceFilePaths[i] = filepath.Join(
						sourcesFullPath,
						sourceFileNames[i],
					)

					err = ioutil.WriteFile(
						sourceFilePaths[i],
						[]byte("some content"),
						os.ModePerm,
					)
					Expect(err).ShouldNot(HaveOccurred())

					remotePaths[i] = fmt.Sprintf(
						"product_files/%s/%s",
						s3FilepathPrefix,
						sourceFileNames[i],
					)
				}

				outRequest.Params.FileGlob = fmt.Sprintf("%s/*", sourcesDir)
				outRequest.Params.FilepathPrefix = s3FilepathPrefix

				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				By("Removing uploaded file")
				for i := 0; i < totalFiles; i++ {
					client.DeleteFile(pivnetBucketName, remotePaths[i])
				}
			})

			It("uploads files to s3 and creates files on pivnet", func() {
				By("Getting existing list of product files")
				existingProductFiles := getProductFiles(productSlug)

				By("Verifying existing product files does not yet contain new files")
				var existingProductFileNames []string
				for _, f := range existingProductFiles {
					existingProductFileNames = append(existingProductFileNames, f.Name)
				}
				for i := 0; i < totalFiles; i++ {
					Expect(existingProductFileNames).NotTo(ContainElement(sourceFileNames[i]))
				}

				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, executableTimeout).Should(gexec.Exit(0))

				By("Verifying uploaded files can be downloaded")
				for i := 0; i < totalFiles; i++ {
					localDownloadPath := fmt.Sprintf("%s-downloaded", sourceFilePaths[i])
					err := client.DownloadFile(pivnetBucketName, remotePaths[i], localDownloadPath)
					Expect(err).ShouldNot(HaveOccurred())
				}

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(response.Version.ProductVersion).To(Equal(productVersion))

				By("Getting updated list of product files")
				updatedProductFiles := getProductFiles(productSlug)

				By("Verifying number of product files has increased by the expected amount")
				newProductFileCount := len(updatedProductFiles) - len(existingProductFiles)
				Expect(newProductFileCount).To(Equal(totalFiles))

				By("Verifying updated product files contains new files")
				var newProductFiles []pivnet.ProductFile
				for _, p := range updatedProductFiles {
					if stringInSlice(p.Name, sourceFileNames) {
						newProductFiles = append(newProductFiles, p)
					}
				}
				Expect(len(newProductFiles)).To(Equal(totalFiles))

				By("Getting newly-created release")
				release, err := pivnetClient.GetRelease(productSlug, productVersion)
				Expect(err).ShouldNot(HaveOccurred())

				By("Verifying release contains new product files")
				productFilesFromRelease, err := pivnetClient.GetProductFiles(release)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(productFilesFromRelease.ProductFiles)).To(Equal(totalFiles))
				for _, p := range productFilesFromRelease.ProductFiles {
					Expect(sourceFileNames).To(ContainElement(p.Name))
				}

				By("Downloading the files via the In CMD")
				inRequest := concourse.InRequest{
					Source: concourse.Source{
						APIToken:    pivnetAPIToken,
						ProductSlug: productSlug,
					},
					Version: concourse.Version{
						ProductVersion: productVersion,
					},
				}

				destDirectory, err := ioutil.TempDir("", "pivnet-out-test")
				Expect(err).NotTo(HaveOccurred())

				stdinContents, err = json.Marshal(inRequest)
				Expect(err).NotTo(HaveOccurred())

				downloadCmd := exec.Command(inPath, destDirectory)

				downloadSession := run(downloadCmd, stdinContents)
				Eventually(downloadSession, executableTimeout).Should(gexec.Exit(0))

				By("Deleting created files on pivnet")
				for _, p := range newProductFiles {
					_, err := pivnetClient.DeleteProductFile(productSlug, p.ID)
					Expect(err).ShouldNot(HaveOccurred())
				}
			})
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
				outRequest.Params.FileGlob = ""
				outRequest.Params.FilepathPrefix = ""

				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating the new product version does not yet exist")
				productVersions := getProductVersions(productSlug)
				Expect(productVersions).NotTo(BeEmpty())
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
