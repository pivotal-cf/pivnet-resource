package acceptance

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

var _ = Describe("In", func() {
	var (
		productSlug    = "pivotal-diego-pcf"
		productVersion = "pivnet-testing"
		destDirectory  string

		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		var err error

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, destDirectory)

		By("Creating default request")
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    pivnetAPIToken,
				ProductSlug: productSlug,
			},
			Version: concourse.Version{
				ProductVersion: productVersion,
			},
		}

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing temporary destination directory")
		err := os.RemoveAll(destDirectory)
		Expect(err).NotTo(HaveOccurred())
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
		files, err := dataDir.Readdir(2)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating files have non-zero-length content")
		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
			Expect(f.Size()).ToNot(BeZero())
		}

		By("Validating filenames are correct")
		Expect(fileNames).To(ContainElement("setup.ps1"))
	})

	It("creates a version file with the downloaded version", func() {
		versionFilepath := filepath.Join(destDirectory, "version")

		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, executableTimeout).Should(gexec.Exit(0))

		By("Validating version file has correct contents")
		contents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(contents)).To(Equal(productVersion))
	})

	Context("when globs are provided", func() {
		It("downloads only the files that match the glob", func() {

			By("setting the glob")
			inRequest.Source.ProductSlug = "p-data-sync"
			inRequest.Version.ProductVersion = "1.1.2.0"
			inRequest.Params.Globs = []string{"*PCFData-1.1.0.a*"}

			globStdInRequest, err := json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())

			By("Running the command")
			session := run(command, globStdInRequest)
			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			By("Reading downloaded files")
			dataDir, err := os.Open(destDirectory)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating number of downloaded files")
			files, err := dataDir.Readdir(6)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(files).To(HaveLen(2))

			By("Validating files have non-zero-length content")
			var fileNames []string
			for _, f := range files {
				fileNames = append(fileNames, f.Name())
				Expect(f.Size()).ToNot(BeZero())
			}

			By("Validating filenames are correct")
			Expect(fileNames).To(ContainElement("PCFData-1.1.0.aar"))
		})
	})
})
