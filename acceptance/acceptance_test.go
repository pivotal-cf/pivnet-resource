package acceptance

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

const (
	checkTimeout = 5 * time.Second
)

var _ = Describe("Acceptance", func() {
	Context("Check", func() {
		Context("when a version is provided", func() {
			It("returns the next version", func() {
				productName := "p-mysql"
				releases := getProductReleases(productName)

				command := exec.Command(checkPath)
				writer, err := command.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				raw, err := json.Marshal(concourse.Request{
					Source: concourse.Source{
						APIToken:     os.Getenv("API_TOKEN"),
						ResourceName: productName,
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

				Expect(response[0].Version).To(Equal(releases[2]))
				Expect(response[1].Version).To(Equal(releases[1]))
				Expect(response[2].Version).To(Equal(releases[0]))
			})
		})
	})
})
