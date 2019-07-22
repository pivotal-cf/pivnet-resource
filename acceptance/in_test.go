package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/versions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		eulaSlug    = "pivotal_beta_eula"
		releaseType = pivnet.ReleaseType("Minor Release")

		version                string
		versionWithFingerprint string
		destDirectory          string

		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
		err           error
	)

	BeforeEach(func() {
		By("Generating 'random' product version")
		version = fmt.Sprintf("%d", time.Now().Nanosecond())

		By("Creating new release")
		release, err := pivnetClient.CreateRelease(pivnet.CreateReleaseConfig{
			ProductSlug: productSlug,
			Version:     version,
			EULASlug:    eulaSlug,
			ReleaseType: string(releaseType),
		})
		Expect(err).NotTo(HaveOccurred())

		versionWithFingerprint, err = versions.CombineVersionAndFingerprint(release.Version, release.SoftwareFilesUpdatedAt)
		Expect(err).NotTo(HaveOccurred())

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, filepath.Join(destDirectory, "my-resource"))

	})

	AfterEach(func() {
		By("Removing temporary destination directory")
		err := os.RemoveAll(destDirectory)
		Expect(err).NotTo(HaveOccurred())

		// We do not delete the release as it causes race conditions with other tests
	})

	Describe("verbose flag", func() {
		BeforeEach(func() {
			By("Creating default request")
			inRequest = concourse.InRequest{
				Source: concourse.Source{
					APIToken:    refreshToken,
					ProductSlug: productSlug,
					Endpoint:    endpoint,
					Verbose:     false,
				},
				Version: concourse.Version{
					ProductVersion: versionWithFingerprint,
				},
			}
		})

		JustBeforeEach(func() {
			stdinContents, err = json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when user does not specify verbose output", func() {
			It("does not print verbose output", func() {
				session := run(command, stdinContents)
				Eventually(session, executableTimeout).Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).NotTo(ContainSubstring("Verbose output enabled"))
			})
		})

		Context("when user specifies verbose output", func() {
			BeforeEach(func() {
				inRequest.Source.Verbose = true
			})

			It("prints verbose output", func() {
				session := run(command, stdinContents)
				Eventually(session, executableTimeout).Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(ContainSubstring("Verbose output enabled"))
			})
		})

	})

	Context("when user supplies UAA refresh token in source config", func() {
		BeforeEach(func() {
			By("Creating default request")
			inRequest = concourse.InRequest{
				Source: concourse.Source{
					APIToken:    refreshToken,
					ProductSlug: productSlug,
					Endpoint:    endpoint,
				},
				Version: concourse.Version{
					ProductVersion: versionWithFingerprint,
				},
			}

			stdinContents, err = json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())
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
			Expect(response.Version.ProductVersion).To(Equal(versionWithFingerprint))

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
				inRequest.Source.APIToken = "iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii"

				var err error
				stdinContents, err = json.Marshal(inRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				By("Running the command")
				session := run(command, stdinContents)

				By("Validating command exited with error")
				Eventually(session, executableTimeout).Should(gexec.Exit())
				Expect(session.Err).Should(gbytes.Say("failed to fetch API token"))
			})
		})
	})
})
