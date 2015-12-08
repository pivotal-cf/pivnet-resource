package acceptance

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Acceptance", func() {
	Context("Check", func() {
		It("can get product versions", func() {
			productName := "p-gitlab"
			currentReleasedVersion := getProductRelease(productName)
			fmt.Println(currentReleasedVersion.Version)

			command := exec.Command(checkPath)
			writer, err := command.StdinPipe()
			Expect(err).ShouldNot(HaveOccurred())

			raw, err := json.Marshal(concourseRequest{
				Source: Source{
					APIToken:     "nada-a-thing",
					ResourceName: productName,
				}})
			Expect(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(writer, string(raw))
			Expect(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})
})
