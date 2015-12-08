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
	"github.com/pivotal-cf-experimental/pivnet-resource"
)

const (
	checkTimeout = 5 * time.Second
)

var _ = Describe("Acceptance", func() {
	Context("Check", func() {
		It("can get product versions", func() {
			productName := "p-gitlab"
			currentRelease := getProductRelease(productName)

			command := exec.Command(checkPath)
			writer, err := command.StdinPipe()
			Expect(err).ShouldNot(HaveOccurred())

			raw, err := json.Marshal(pivnet.ConcourseRequest{
				Source: pivnet.ConcourseSource{
					APIToken:     os.Getenv("API_TOKEN"),
					ResourceName: productName,
				}})
			Expect(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(writer, string(raw))
			Expect(err).ShouldNot(HaveOccurred())

			Eventually(session, checkTimeout).Should(gexec.Exit(0))

			response := pivnet.ConcourseResponse{}
			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(response[0].Version).To(Equal(currentRelease.Version))
		})
	})
})
