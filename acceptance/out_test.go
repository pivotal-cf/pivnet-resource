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
	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/versions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	executableTimeout = 60 * time.Second
)

var _ = Describe("Out", func() {
	var (
		version string

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
		version = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating a metadata struct")
		productMetadata = metadata.Metadata{
			Release: &metadata.Release{
				ReleaseType:     string(releaseType),
				EULASlug:        eulaSlug,
				ReleaseDate:     releaseDate,
				Description:     description,
				ReleaseNotesURL: releaseNotesURL,
				Version:         version,
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

			releaseVersions, err := versionsWithFingerprints(releases)
			Expect(err).NotTo(HaveOccurred())

			Expect(versionsWithoutFingerprints(releaseVersions)).NotTo(ContainElement(version))

			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			By("Validating new release exists on pivnet")
			releases, err = pivnetClient.ReleasesForProductSlug(productSlug)
			Expect(err).NotTo(HaveOccurred())

			releaseVersions, err = versionsWithFingerprints(releases)
			Expect(err).NotTo(HaveOccurred())

			Expect(versionsWithoutFingerprints(releaseVersions)).To(ContainElement(version))

			By("Outputting a valid json response")
			response := concourse.OutResponse{}
			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating the release was created correctly")
			release, err := pivnetClient.GetRelease(productSlug, version)
			Expect(err).NotTo(HaveOccurred())

			expectedVersion, err := versions.CombineVersionAndFingerprint(release.Version, release.UpdatedAt)
			Expect(err).NotTo(HaveOccurred())

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
	})
})

// versionsWithFingerprints adds the release Fingerprints to the release versions
func versionsWithFingerprints(
	releases []pivnet.Release,
) ([]string, error) {
	var allVersions []string
	for _, r := range releases {
		version, err := versions.CombineVersionAndFingerprint(r.Version, r.UpdatedAt)
		if err != nil {
			return nil, err
		}

		allVersions = append(allVersions, version)
	}

	return allVersions, nil
}
