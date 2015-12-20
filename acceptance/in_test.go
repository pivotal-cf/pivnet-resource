package acceptance

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

var _ = Describe("Acceptance", func() {
	var (
		productName    string
		productVersion string
		destDirectory  string
	)

	BeforeEach(func() {
		var err error
		productName = "push-notification-service"
		productVersion = "1.4.0"
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(destDirectory)
	})

	Context("In", func() {
		It("successfully downloads all of the files in the specified release", func() {
			command := exec.Command(inPath, destDirectory)
			writer, err := command.StdinPipe()
			Expect(err).ShouldNot(HaveOccurred())

			raw, err := json.Marshal(concourse.Request{
				Source: concourse.Source{
					APIToken:    os.Getenv("API_TOKEN"),
					ProductName: productName,
				},
				Version: map[string]string{
					"product_version": productVersion,
				},
			})
			Expect(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(writer, string(raw))
			Expect(err).ShouldNot(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))

			dataDir, err := os.Open(destDirectory)
			Expect(err).ShouldNot(HaveOccurred())

			files, err := dataDir.Readdir(2)
			Expect(err).ShouldNot(HaveOccurred())

			var fileNames []string
			for _, f := range files {
				fileNames = append(fileNames, f.Name())
				Expect(f.Size()).ToNot(BeZero())
			}

			Expect(fileNames).To(ConsistOf([]string{"PCFPush-1.4.0.aar", "PCFPush-1.4.0.framework.zip"}))
		})
	})
})
