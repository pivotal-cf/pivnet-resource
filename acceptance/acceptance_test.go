package acceptance

import (
	"encoding/json"
	"io"
	"os"
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

var _ = Describe("Acceptance", func() {
	Context("Check", func() {
		var releases []string
		productName := "p-mysql"
		BeforeEach(func() {
			releases = getProductReleases(productName)
		})

		Context("when no version is provided", func() {
			It("returns the most recent version", func() {
				command := exec.Command(checkPath)
				writer, err := command.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				raw, err := json.Marshal(concourse.Request{
					Source: concourse.Source{
						APIToken:    os.Getenv("API_TOKEN"),
						ProductName: productName,
					},
				})
				Expect(err).ShouldNot(HaveOccurred())

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(writer, string(raw))
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(session, checkTimeout).Should(gexec.Exit(0))

				response := concourse.Response{}
				err = json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response).To(HaveLen(1))

				Expect(response[0].ProductVersion).To(Equal(releases[0]))
			})
		})

		Context("when a version is provided", func() {
			It("returns all newer versions", func() {
				command := exec.Command(checkPath)
				writer, err := command.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				raw, err := json.Marshal(concourse.Request{
					Source: concourse.Source{
						APIToken:    os.Getenv("API_TOKEN"),
						ProductName: productName,
					},
					Version: map[string]string{
						"version": releases[3],
					},
				})
				Expect(err).ShouldNot(HaveOccurred())

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(writer, string(raw))
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(session, checkTimeout).Should(gexec.Exit(0))

				response := concourse.Response{}
				err = json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response).To(HaveLen(3))

				Expect(response[0].ProductVersion).To(Equal(releases[2]))
				Expect(response[1].ProductVersion).To(Equal(releases[1]))
				Expect(response[2].ProductVersion).To(Equal(releases[0]))
			})
		})

		Context("when no api_token is provided", func() {
			It("exits with error", func() {
				command := exec.Command(checkPath)
				writer, err := command.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				raw, err := json.Marshal(concourse.Request{
					Source: concourse.Source{
						APIToken:    "",
						ProductName: productName,
					},
					Version: map[string]string{
						"version": releases[0],
					},
				})
				Expect(err).ShouldNot(HaveOccurred())

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(writer, string(raw))
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(session, checkTimeout).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("api_token must be provided"))
			})
		})

		Context("when no product_name is provided", func() {
			It("exits with error", func() {
				command := exec.Command(checkPath)
				writer, err := command.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				raw, err := json.Marshal(concourse.Request{
					Source: concourse.Source{
						APIToken:    "some-api-token",
						ProductName: "",
					},
					Version: map[string]string{
						"version": releases[0],
					},
				})
				Expect(err).ShouldNot(HaveOccurred())

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(writer, string(raw))
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(session, checkTimeout).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("product_name must be provided"))
			})
		})
	})
})
