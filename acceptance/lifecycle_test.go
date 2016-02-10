package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

var _ = Describe("Lifecycle test", func() {
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

		filePrefix = "pivnet-resource-test-file"

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

	Describe("Creating a new release", func() {
		AfterEach(func() {
			By("Deleting newly-created release")
			deletePivnetRelease(productSlug, productVersion)
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
						"%s-%d",
						filePrefix,
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

				By("Verifying uploaded files can be downloaded directly from S3")
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

					productFile, err := pivnetClient.GetProductFile(productSlug, p.ID)
					Expect(err).ShouldNot(HaveOccurred())
					// Contents are fixed at 'some contents' so the MD5 is known.
					Expect(productFile.MD5).To(Equal("9893532233caff98cd083a116b013c0b"))
				}

				By("Downloading all files via in command and glob")
				inRequest := concourse.InRequest{
					Source: concourse.Source{
						APIToken:    pivnetAPIToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Params: concourse.InParams{
						Globs: []string{"*"},
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

				By("Running the command")
				inSession := run(downloadCmd, stdinContents)
				Eventually(inSession, executableTimeout).Should(gexec.Exit(0))

				By("Validating number of downloaded files")
				files, err := ioutil.ReadDir(destDirectory)
				Expect(err).ShouldNot(HaveOccurred())

				expectedFileCount := totalFiles + 1 // the version file will be present.
				Expect(err).ShouldNot(HaveOccurred())
				Expect(files).To(HaveLen(expectedFileCount))

				By("Validating files have non-zero-length content")
				for _, f := range files {
					Expect(f.Size()).To(BeNumerically(">", 0))
				}

				By("Downloading no files via in command and no glob")
				inRequest = concourse.InRequest{
					Source: concourse.Source{
						APIToken:    pivnetAPIToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Version: concourse.Version{
						ProductVersion: productVersion,
					},
				}

				destDirectory, err = ioutil.TempDir("", "pivnet-out-test")
				Expect(err).NotTo(HaveOccurred())

				stdinContents, err = json.Marshal(inRequest)
				Expect(err).NotTo(HaveOccurred())

				downloadCmd = exec.Command(inPath, destDirectory)

				By("Running the command")
				inSession = run(downloadCmd, stdinContents)
				Eventually(inSession, executableTimeout).Should(gexec.Exit(0))

				By("Validating number of downloaded files")
				files, err = ioutil.ReadDir(destDirectory)
				Expect(err).ShouldNot(HaveOccurred())

				expectedFileCount = 1 // the version file will be present.
				Expect(err).ShouldNot(HaveOccurred())
				Expect(files).To(HaveLen(expectedFileCount))

				Expect(files[0].Name()).To(Equal("version"))

				By("Expecting error with in command and mismatched globs")
				inRequest = concourse.InRequest{
					Source: concourse.Source{
						APIToken:    pivnetAPIToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Params: concourse.InParams{
						Globs: []string{filePrefix + "*", "badglob"},
					},
					Version: concourse.Version{
						ProductVersion: productVersion,
					},
				}

				destDirectory, err = ioutil.TempDir("", "pivnet-out-test")
				Expect(err).NotTo(HaveOccurred())

				stdinContents, err = json.Marshal(inRequest)
				Expect(err).NotTo(HaveOccurred())

				downloadCmd = exec.Command(inPath, destDirectory)

				By("Running the command, expecting error")
				inSession = run(downloadCmd, stdinContents)
				Eventually(inSession, executableTimeout).Should(gexec.Exit(1))

				By("Deleting created files on pivnet")
				for _, p := range newProductFiles {
					_, err := pivnetClient.DeleteProductFile(productSlug, p.ID)
					Expect(err).ShouldNot(HaveOccurred())
				}
			})
		})
	})
})
