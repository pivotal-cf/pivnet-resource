package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/versions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lifecycle test", func() {
	var (
		releaseType     = "Minor Release"
		releaseDate     = "2015-12-17"
		eulaSlug        = "pivotal_beta_eula"
		description     = "this release is for automated-testing only."
		releaseNotesURL = "https://example.com"

		metadataFile =  "metadata"
		metadataFile2 = "metadata2"
		version      string
		version2     string

		filePrefix = "pivnet-resource-test-file"

		command       *exec.Cmd
		command2      *exec.Cmd
		stdinContents  []byte
		stdinContents2 []byte
		outRequest    concourse.OutRequest
		outRequest2   concourse.OutRequest
		rootDir       string
	)

	additionalSynchronizedBeforeSuite = func(suiteEnv SuiteEnv) {
		By("Clean up product-files")

		By("Get product files")
		productFiles, err := pivnetClient.ProductFiles(suiteEnv.ProductSlug)
		Expect(err).NotTo(HaveOccurred())

		By("Deleting created files on pivnet")
		for _, p := range productFiles {
			if strings.Contains(p.Name, filePrefix) {
				_, err := pivnetClient.DeleteProductFile(suiteEnv.ProductSlug, p.ID)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}

		By("Deleting image references on pivnet")
		imageRefs, err := pivnetClient.ImageReferences(suiteEnv.ProductSlug)
		Expect(err).ShouldNot(HaveOccurred())
		for _, i := range imageRefs {
			_, err := pivnetClient.DeleteImageReference(suiteEnv.ProductSlug, i.ID)
			Expect(err).ShouldNot(HaveOccurred())
		}
	}

	BeforeEach(func() {
		var err error

		By("Creating a temporary root dir")
		rootDir, err = ioutil.TempDir("", "")
		Expect(err).ShouldNot(HaveOccurred())

		By("Generating 'random' product version")
		version = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating a metadata struct")
		productMetadata := metadata.Metadata{
			Release: &metadata.Release{
				ReleaseType:     releaseType,
				EULASlug:        eulaSlug,
				ReleaseDate:     releaseDate,
				Description:     description,
				ReleaseNotesURL: releaseNotesURL,
				Version:         version,
			},
			ImageReferences: []metadata.ImageReference{
				{
					Name:      imageName,
					ImagePath: imagePath,
					Digest:    imageDigest,
				},
			},
		}

		version2 = fmt.Sprintf("%d", time.Now().Nanosecond())

		productMetadata2 := metadata.Metadata{
			Release: &metadata.Release{
				ReleaseType:     releaseType,
				EULASlug:        eulaSlug,
				ReleaseDate:     releaseDate,
				Description:     description,
				ReleaseNotesURL: releaseNotesURL,
				Version:         version2,
			},
		}


		By("Marshaling the metadata to yaml")
		metadataBytes, err := yaml.Marshal(productMetadata)
		Expect(err).ShouldNot(HaveOccurred())
		metadataBytes2, err := yaml.Marshal(productMetadata2)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing the metadata to a file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, metadataFile),
			metadataBytes,
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		err = ioutil.WriteFile(
			filepath.Join(rootDir, metadataFile2),
			metadataBytes2,
			os.ModePerm)
		Expect(err).ShouldNot(HaveOccurred())

		By("Creating command object")
		command = exec.Command(outPath, rootDir)
		command2 = exec.Command(outPath, rootDir)

		By("Creating default request")
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        refreshToken,
				ProductSlug:     productSlug,
				Endpoint:        endpoint,
			},
			Params: concourse.OutParams{
				FileGlob:       "*",
				MetadataFile:   metadataFile,
			},
		}

		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())

		outRequest2 = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        refreshToken,
				ProductSlug:     productSlug,
				Endpoint:        endpoint,
			},
			Params: concourse.OutParams{
				FileGlob:       "*",
				MetadataFile:   metadataFile2,
			},
		}

		stdinContents2, err = json.Marshal(outRequest2)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing local temp files")
		err := os.RemoveAll(rootDir)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Creating a new release", func() {
		// We do not delete the release as it causes race conditions with other tests

		Context("when S3 source and params are configured correctly", func() {
			var (
				sourcesDir      = "sources"
				sourceFileNames []string
				sourceFilePaths []string
				remotePaths     []string

				totalFiles = 3
			)

			BeforeEach(func() {
				var err error

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
						"%s",
						sourceFileNames[i],
					)
				}

				outRequest.Params.FileGlob = fmt.Sprintf("%s/*", sourcesDir)

				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())

				outRequest2.Params.FileGlob = fmt.Sprintf("%s/*", sourcesDir)

				stdinContents2, err = json.Marshal(outRequest2)
				Expect(err).ShouldNot(HaveOccurred())

			})

			It("uploads files to s3 and creates files on pivnet", func() {
				By("Getting existing list of product files")
				existingProductFiles, err := pivnetClient.ProductFiles(productSlug)
				Expect(err).NotTo(HaveOccurred())

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
				Eventually(session, 10 * time.Minute).Should(gexec.Exit(0))

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err = json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating the release was created correctly")
				release, err := pivnetClient.GetRelease(productSlug, version)
				Expect(err).NotTo(HaveOccurred())

				expectedVersion, err := versions.CombineVersionAndFingerprint(release.Version, release.SoftwareFilesUpdatedAt)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Version.ProductVersion).To(Equal(expectedVersion))

				By("Getting updated list of product files")
				updatedProductFiles, err := pivnetClient.ProductFiles(productSlug)
				Expect(err).NotTo(HaveOccurred())

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
				release, err = pivnetClient.GetRelease(productSlug, version)
				Expect(err).ShouldNot(HaveOccurred())

				expectedVersionWithFingerprint, err := versions.CombineVersionAndFingerprint(release.Version, release.SoftwareFilesUpdatedAt)
				Expect(err).NotTo(HaveOccurred())

				By("Verifying release contains new product files")
				productFilesFromRelease, err := pivnetClient.ProductFilesForRelease(productSlug, release.ID)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(productFilesFromRelease)).To(Equal(totalFiles))
				for _, p := range productFilesFromRelease {
					Expect(sourceFileNames).To(ContainElement(p.Name))

					productFile, err := pivnetClient.ProductFileForRelease(
						productSlug,
						release.ID,
						p.ID,
					)
					Expect(err).ShouldNot(HaveOccurred())
					// Contents are fixed at 'some contents'
					Expect(productFile.SHA256).To(Equal("290f493c44f5d63d06b374d0a5abd292fae38b92cab2fae5efefe1b0e9347f56"))
				}

				By("Running a new command to create a second release with the same product files")
				session2 := run(command2, stdinContents2)
				Eventually(session2, executableTimeout).Should(gexec.Exit(0))

				By("Outputting a valid json response for the second release")
				response2 := concourse.OutResponse{}
				err = json.Unmarshal(session2.Out.Contents(), &response2)
				Expect(err).ShouldNot(HaveOccurred())

				By("Getting the newer release")
				releaseWithExistingProductFiles, err := pivnetClient.GetRelease(productSlug, version2)
				Expect(err).ShouldNot(HaveOccurred())

				expectedVersionWithFingerprint2, err := versions.CombineVersionAndFingerprint(releaseWithExistingProductFiles.Version, releaseWithExistingProductFiles.SoftwareFilesUpdatedAt)
				Expect(err).NotTo(HaveOccurred())

				By("Validating that the newer release was created correctly")
				Expect(response2.Version.ProductVersion).To(Equal(expectedVersionWithFingerprint2))

				By("Getting the updated list of product files for second release")
				updatedProductFiles2, err := pivnetClient.ProductFiles(productSlug)
				Expect(err).NotTo(HaveOccurred())

				By("Verifying that the number of product files has not increased")
				numProductFilesAdded := len(updatedProductFiles2) - len(updatedProductFiles)
				Expect(numProductFilesAdded).To(Equal(0))

				By("Verifying that the newer release contains existing product files")
				productFilesFromRelease2, err := pivnetClient.ProductFilesForRelease(productSlug, releaseWithExistingProductFiles.ID)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(productFilesFromRelease2)).To(Equal(totalFiles))
				for _, p := range productFilesFromRelease2 {
					Expect(sourceFileNames).To(ContainElement(p.Name))

					productFile, err := pivnetClient.ProductFileForRelease(
						productSlug,
						releaseWithExistingProductFiles.ID,
						p.ID,
					)
					Expect(err).ShouldNot(HaveOccurred())
					// Contents are fixed at 'some contents'
					Expect(productFile.SHA256).To(Equal("290f493c44f5d63d06b374d0a5abd292fae38b92cab2fae5efefe1b0e9347f56"))
				}

				By("Verifying that the product files are still contained in the older release")
				productFilesFromRelease, err = pivnetClient.ProductFilesForRelease(productSlug, release.ID)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(productFilesFromRelease)).To(Equal(totalFiles))
				for _, p := range productFilesFromRelease {
					Expect(sourceFileNames).To(ContainElement(p.Name))

					productFile, err := pivnetClient.ProductFileForRelease(
						productSlug,
						release.ID,
						p.ID,
					)
					Expect(err).ShouldNot(HaveOccurred())
					// Contents are fixed at 'some contents'
					Expect(productFile.SHA256).To(Equal("290f493c44f5d63d06b374d0a5abd292fae38b92cab2fae5efefe1b0e9347f56"))
				}

				By("Downloading all files via in command and glob")
				inRequest := concourse.InRequest{
					Source: concourse.Source{
						APIToken:    refreshToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Params: concourse.InParams{
						Globs: []string{"*"},
					},
					Version: concourse.Version{
						ProductVersion: expectedVersionWithFingerprint,
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

				// one file is version; two files are metadata
				expectedFileCount := totalFiles + 3
				Expect(err).ShouldNot(HaveOccurred())
				Expect(files).To(HaveLen(expectedFileCount))

				By("Validating files have non-zero-length content")
				for _, f := range files {
					Expect(f.Size()).To(BeNumerically(">", 0))
				}

				By("Downloading no files via in command and no glob")
				inRequest = concourse.InRequest{
					Source: concourse.Source{
						APIToken:    refreshToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Version: concourse.Version{
						ProductVersion: expectedVersionWithFingerprint,
					},
					Params: concourse.InParams{
						Globs: []string{},
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

				// one file is version; two files are metadata
				expectedFileCount = 3
				Expect(err).ShouldNot(HaveOccurred())
				Expect(files).To(HaveLen(expectedFileCount))

				Expect(files[0].Name()).To(Equal("metadata.json"))
				Expect(files[1].Name()).To(Equal("metadata.yaml"))
				Expect(files[2].Name()).To(Equal("version"))

				By("Expecting error with in command and mismatched globs")
				inRequest = concourse.InRequest{
					Source: concourse.Source{
						APIToken:    refreshToken,
						ProductSlug: productSlug,
						Endpoint:    endpoint,
					},
					Params: concourse.InParams{
						Globs: []string{"badglob"},
					},
					Version: concourse.Version{
						ProductVersion: expectedVersionWithFingerprint,
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
