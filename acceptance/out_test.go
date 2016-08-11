package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	executableTimeout = 60 * time.Second
)

var _ = Describe("Out", func() {
	var (
		productVersion string

		releaseType     = pivnet.ReleaseType("Minor Release")
		releaseDate     = "2015-12-17"
		eulaSlug        = "pivotal_beta_eula"
		description     = "this release is for automated-testing only."
		releaseNotesURL = "https://example.com"

		metadataFile = "metadata"

		command         *exec.Cmd
		stdinContents   []byte
		outRequest      concourse.OutRequest
		rootDir         string
		productMetadata metadata.Metadata
	)

	BeforeEach(func() {
		var err error

		By("Creating a temporary root dir")
		rootDir, err = ioutil.TempDir("", "")
		Expect(err).ShouldNot(HaveOccurred())

		By("Generating 'random' product version")
		productVersion = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating a metadata struct")
		productMetadata = metadata.Metadata{
			Release: &metadata.Release{
				ReleaseType:     string(releaseType),
				EULASlug:        eulaSlug,
				ReleaseDate:     releaseDate,
				Description:     description,
				ReleaseNotesURL: releaseNotesURL,
				Version:         productVersion,
			},
		}

		By("Marshaling the metadata to yaml")
		metadataBytes, err := yaml.Marshal(productMetadata)
		Expect(err).ShouldNot(HaveOccurred())

		By("Writing the metadata to a file")
		err = ioutil.WriteFile(
			filepath.Join(rootDir, metadataFile),
			metadataBytes,
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
				FileGlob:       "",
				FilepathPrefix: "",
				MetadataFile:   metadataFile,
			},
		}
	})

	JustBeforeEach(func() {
		var err error
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
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("usage"))
			})
		})

		Context("when metadata file value is empty", func() {
			BeforeEach(func() {
				outRequest.Params.MetadataFile = ""
			})

			It("exits with error", func() {
				session := run(command, stdinContents)

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("metadata_file"))
			})
		})
	})

	Describe("Creating a new release", func() {
		// We do not delete the release as it causes race conditions with other tests

		It("Successfully creates a release", func() {
			var err error
			stdinContents, err = json.Marshal(outRequest)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the new product version does not yet exist")
			releases, err := pivnetClient.ReleasesForProductSlug(productSlug)
			Expect(err).NotTo(HaveOccurred())

			productVersions, err := pivnetClient.ProductVersions(productSlug, releases)
			Expect(err).NotTo(HaveOccurred())

			Expect(productVersionsWithoutETags(productVersions)).NotTo(ContainElement(productVersion))

			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			By("Validating new release exists on pivnet")
			releases, err = pivnetClient.ReleasesForProductSlug(productSlug)
			Expect(err).NotTo(HaveOccurred())

			productVersions, err = pivnetClient.ProductVersions(productSlug, releases)
			Expect(err).NotTo(HaveOccurred())

			Expect(productVersionsWithoutETags(productVersions)).To(ContainElement(productVersion))

			By("Outputting a valid json response")
			response := concourse.OutResponse{}
			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the release was created correctly")
			release, err := pivnetClient.GetRelease(productSlug, productVersion)
			Expect(err).NotTo(HaveOccurred())

			releaseETag, err := pivnetClient.ReleaseETag(productSlug, release.ID)
			Expect(err).NotTo(HaveOccurred())

			expectedVersion := fmt.Sprintf("%s#%s", productVersion, releaseETag)
			Expect(response.Version.ProductVersion).To(Equal(expectedVersion))

			Expect(release.ReleaseType).To(Equal(releaseType))
			Expect(release.ReleaseDate).To(Equal(releaseDate))
			Expect(release.EULA.Slug).To(Equal(eulaSlug))
			Expect(release.Description).To(Equal(description))
			Expect(release.ReleaseNotesURL).To(Equal(releaseNotesURL))

			By("Validing the returned metadata")
			metadataReleaseType, err := metadataValueForKey(response.Metadata, "release_type")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadataReleaseType).To(Equal(string(releaseType)))

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
				availability = "Selected User Groups Only"
				userGroupIDs = []string{"6", "8", "54"}
			)

			BeforeEach(func() {
				By("Adding availability to the metadata")
				productMetadata.Release.Availability = availability
				productMetadata.Release.UserGroupIDs = userGroupIDs

				By("Marshaling the metadata to yaml")
				metadataBytes, err := yaml.Marshal(productMetadata)
				Expect(err).ShouldNot(HaveOccurred())

				By("Writing the metadata to a file")
				err = ioutil.WriteFile(
					filepath.Join(rootDir, metadataFile),
					metadataBytes,
					os.ModePerm)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("Creates a release and updates the availability and user groups", func() {
				var err error
				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating the new product version does not yet exist")
				releases, err := pivnetClient.ReleasesForProductSlug(productSlug)
				Expect(err).NotTo(HaveOccurred())

				productVersions, err := pivnetClient.ProductVersions(productSlug, releases)
				Expect(err).NotTo(HaveOccurred())

				Expect(productVersionsWithoutETags(productVersions)).NotTo(ContainElement(productVersion))

				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, executableTimeout).Should(gexec.Exit(0))

				By("Validating new release exists on pivnet")
				releases, err = pivnetClient.ReleasesForProductSlug(productSlug)
				Expect(err).NotTo(HaveOccurred())

				productVersions, err = pivnetClient.ProductVersions(productSlug, releases)
				Expect(err).NotTo(HaveOccurred())

				Expect(productVersionsWithoutETags(productVersions)).To(ContainElement(productVersion))

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err = json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(response.Version.ProductVersion).To(ContainSubstring(productVersion))

				By("Validating the release was created correctly")
				release, err := pivnetClient.GetRelease(productSlug, productVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(release.Availability).To(Equal(availability))

				By("Validing the returned metadata")
				metadataAvailability, err := metadataValueForKey(response.Metadata, "availability")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(metadataAvailability).To(Equal(availability))

				By("Validating the user groups were associated with the release")
				userGroups, err := pivnetClient.UserGroups(productSlug, release.ID)
				Expect(err).NotTo(HaveOccurred())

				userGroupIDs := []int{}
				for _, userGroup := range userGroups {
					userGroupIDs = append(userGroupIDs, userGroup.ID)
				}
				Expect(userGroupIDs).Should(ConsistOf(6, 8, 54))
			})
		})
	})
})
