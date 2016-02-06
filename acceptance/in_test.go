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

var _ = Describe("In", func() {
	var (
		eulaSlug    = "pivotal_beta_eula"
		releaseType = "Minor Release"

		release        pivnet.Release
		productVersion string
		destDirectory  string

		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		var err error

		By("Generating 'random' product version")
		productVersion = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating new version")
		release, err = pivnetClient.CreateRelease(pivnet.CreateReleaseConfig{
			ProductSlug:    productSlug,
			ProductVersion: productVersion,
			EulaSlug:       eulaSlug,
			ReleaseType:    releaseType,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, destDirectory)

		By("Creating default request")
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

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing temporary destination directory")
		err := os.RemoveAll(destDirectory)
		Expect(err).NotTo(HaveOccurred())

		By("Deleting newly-created release")
		deletePivnetRelease(productSlug, productVersion)
	})

	It("returns valid json", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Outputting a valid json response")
		response := concourse.InResponse{}
		err := json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating output contains correct product version")
		Expect(response.Version.ProductVersion).To(Equal(productVersion))

		By("Validing the returned metadata is present")
		_, err = metadataValueForKey(response.Metadata, "release_type")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = metadataValueForKey(response.Metadata, "release_date")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = metadataValueForKey(response.Metadata, "description")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = metadataValueForKey(response.Metadata, "release_notes_url")
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("does not download any of the files in the specified release", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Validating number of downloaded files is zero")
		files, err := ioutil.ReadDir(destDirectory)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(files)).To(Equal(1)) // the version file will always exist
	})

	It("creates a version file with the downloaded version", func() {
		versionFilepath := filepath.Join(destDirectory, "version")

		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Validating version file has correct contents")
		contents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(contents)).To(Equal(productVersion))
	})
})
