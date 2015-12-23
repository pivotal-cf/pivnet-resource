package acceptance

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

var _ = Describe("In", func() {
	var (
		productName    = "pivotal-diego-pcf"
		productVersion = "pivnet-testing"
		destDirectory  string

		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		var err error
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, destDirectory)

		By("Creating default request")
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    pivnetAPIToken,
				ProductName: productName,
			},
			Version: concourse.Release{
				ProductVersion: productVersion,
			},
		}

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing temporary destination directory")
		os.RemoveAll(destDirectory)
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
	})

	It("successfully downloads all of the files in the specified release", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Reading downloaded files")
		dataDir, err := os.Open(destDirectory)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating number of downloaded files")
		files, err := dataDir.Readdir(1)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating files have non-zero-length content")
		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
			Expect(f.Size()).ToNot(BeZero())
		}

		By("Validating filenames are correct")
		Expect(fileNames).To(ConsistOf([]string{"setup.ps1"}))
	})
})
