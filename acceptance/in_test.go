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
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("In", func() {
	var (
		eulaSlug    = "pivotal_beta_eula"
		releaseType = "Minor Release"

		productVersion  string
		etag            string
		versionWithETag string
		destDirectory   string

		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Generating 'random' product version")
		productVersion = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating new release")
		release, err := pivnetClient.CreateRelease(pivnet.CreateReleaseConfig{
			ProductSlug:    productSlug,
			ProductVersion: productVersion,
			EULASlug:       eulaSlug,
			ReleaseType:    releaseType,
		})
		Expect(err).NotTo(HaveOccurred())

		etag, err = pivnetClient.ReleaseETag(productSlug, release)
		Expect(err).NotTo(HaveOccurred())

		versionWithETag, err = versions.CombineVersionAndETag(productVersion, etag)
		Expect(err).NotTo(HaveOccurred())

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, filepath.Join(destDirectory, "my-resource"))

		By("Creating default request")
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    pivnetAPIToken,
				ProductSlug: productSlug,
				Endpoint:    endpoint,
			},
			Version: concourse.Version{
				ProductVersion: versionWithETag,
			},
		}

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing temporary destination directory")
		err := os.RemoveAll(destDirectory)
		Expect(err).NotTo(HaveOccurred())

		// We do not delete the release as it causes race conditions with other tests
	})

	It("returns valid json", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Printing what is actually happening")
		Eventually(session.Err).Should(gbytes.Say("Getting release"))
		Eventually(session.Err).Should(gbytes.Say("Writing metadata to json file"))

		By("Outputting a valid json response")
		response := concourse.InResponse{}
		err := json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating output contains correct product version")
		Expect(response.Version.ProductVersion).To(Equal(versionWithETag))

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

	Context("when validation fails", func() {
		BeforeEach(func() {
			inRequest.Source.APIToken = ""

			var err error
			stdinContents, err = json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, executableTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say("api_token must be provided"))
		})
	})
})
