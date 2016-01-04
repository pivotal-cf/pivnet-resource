package acceptance

import (
	"encoding/json"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

const (
	checkTimeout = 5 * time.Second
)

var _ = Describe("Check", func() {
	var (
		productSlug = "p-mysql"

		releases      []string
		command       *exec.Cmd
		checkRequest  concourse.CheckRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Getting expected releases")
		releases = getProductVersions(productSlug)

		By("Creating command object")
		command = exec.Command(checkPath)

		By("Creating default request")
		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    pivnetAPIToken,
				ProductSlug: productSlug,
			},
			Version: concourse.Version{
				ProductVersion: releases[3],
			},
		}

		var err error
		stdinContents, err = json.Marshal(checkRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("when no version is provided", func() {
		BeforeEach(func() {
			checkRequest.Version = concourse.Version{}

			var err error
			stdinContents, err = json.Marshal(checkRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("returns the most recent version", func() {
			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, checkTimeout).Should(gexec.Exit(0))

			By("Outputting a valid json response")
			response := concourse.CheckResponse{}
			err := json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating exactly one element was returned")
			Expect(response).To(HaveLen(1))

			By("Validating the returned element has the expected product version")
			Expect(response[0].ProductVersion).To(Equal(releases[0]))
		})
	})

	Context("when a version is provided", func() {
		It("returns all newer versions", func() {
			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, checkTimeout).Should(gexec.Exit(0))

			By("Outputting a valid json response")
			response := concourse.CheckResponse{}
			err := json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating all the expected elements were returned")
			Expect(response).To(HaveLen(3))

			By("Validating the returned elements have the expected product versions")
			Expect(response[0].ProductVersion).To(Equal(releases[2]))
			Expect(response[1].ProductVersion).To(Equal(releases[1]))
			Expect(response[2].ProductVersion).To(Equal(releases[0]))
		})
	})

	Context("when no api_token is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.APIToken = ""

			var err error
			stdinContents, err = json.Marshal(checkRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, checkTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say("api_token must be provided"))
		})
	})

	Context("when no product_slug is provided", func() {
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
			Eventually(session, checkTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say("product_slug must be provided"))
		})
	})
})
