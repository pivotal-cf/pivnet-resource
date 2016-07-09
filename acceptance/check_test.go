package acceptance

import (
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

var _ = Describe("Check", func() {
	var (
		productSlug = "p-mysql"

		versions      []string
		command       *exec.Cmd
		checkRequest  concourse.CheckRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Getting expected releases")
		allReleases, err := pivnetClient.ReleasesForProductSlug(productSlug)
		Expect(err).NotTo(HaveOccurred())

		releases := allReleases[:4]

		By("Getting release ETags")
		versions = make([]string, len(releases))
		for i, r := range releases {
			releaseETag, err := pivnetClient.ReleaseETag(productSlug, r)
			Expect(err).NotTo(HaveOccurred())

			versions[i] = fmt.Sprintf("%s#%s", r.Version, releaseETag)
		}

		By("Creating command object")
		command = exec.Command(checkPath)

		By("Creating default request")
		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    pivnetAPIToken,
				ProductSlug: productSlug,
				Endpoint:    endpoint,
			},
			Version: concourse.Version{
				ProductVersion: versions[3],
			},
		}

		stdinContents, err = json.Marshal(checkRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("returns all newer versions than the provided version without error", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Outputting a valid json response")
		response := concourse.CheckResponse{}
		err := json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating all the expected elements were returned")
		Expect(response).To(HaveLen(3))

		By("Validating the returned elements have the expected product versions")
		Expect(response[0].ProductVersion).To(Equal(versions[2]))
		Expect(response[1].ProductVersion).To(Equal(versions[1]))
		Expect(response[2].ProductVersion).To(Equal(versions[0]))
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			checkRequest.Source.ProductSlug = ""

			var err error
			stdinContents, err = json.Marshal(checkRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, executableTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say("product_slug must be provided"))
		})
	})
})
