package filesystem_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-resource/v2/in/filesystem"
	"github.com/pivotal-cf/pivnet-resource/v2/metadata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileWriter", func() {
	var (
		fileWriter  *filesystem.FileWriter
		downloadDir string
		fakeLogger  logger.Logger
	)

	BeforeEach(func() {
		var err error
		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		fileWriter = filesystem.NewFileWriter(downloadDir, fakeLogger)
	})

	AfterEach(func() {
		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("WriteVersionFile", func() {
		It("writes version file", func() {
			err := fileWriter.WriteVersionFile("some-version")
			Expect(err).NotTo(HaveOccurred())

			expectedVersionFilepath := filepath.Join(downloadDir, "version")
			b, err := ioutil.ReadFile(expectedVersionFilepath)
			Expect(err).NotTo(HaveOccurred())

			Expect(b).To(Equal([]byte("some-version")))
		})
	})

	Describe("WriteMetadataJSONFile", func() {
		It("writes metadata file in json format", func() {
			inputMetadata := metadata.Metadata{
				Release: &metadata.Release{
					Version:     "some version",
					ReleaseType: "some release type",
				},
			}

			err := fileWriter.WriteMetadataJSONFile(inputMetadata)
			Expect(err).NotTo(HaveOccurred())

			expectedVersionFilepath := filepath.Join(downloadDir, "metadata.json")
			b, err := ioutil.ReadFile(expectedVersionFilepath)
			Expect(err).NotTo(HaveOccurred())

			var unmarshalledMetadata metadata.Metadata
			err = json.Unmarshal(b, &unmarshalledMetadata)
			Expect(err).NotTo(HaveOccurred())

			Expect(unmarshalledMetadata).To(Equal(inputMetadata))
		})
	})

	Describe("WriteMetadataYAMLFile", func() {
		It("writes metadata file in yaml format", func() {
			inputMetadata := metadata.Metadata{
				Release: &metadata.Release{
					Version:     "some version",
					ReleaseType: "some release type",
				},
			}

			err := fileWriter.WriteMetadataYAMLFile(inputMetadata)
			Expect(err).NotTo(HaveOccurred())

			expectedVersionFilepath := filepath.Join(downloadDir, "metadata.yaml")
			b, err := ioutil.ReadFile(expectedVersionFilepath)
			Expect(err).NotTo(HaveOccurred())

			var unmarshalledMetadata metadata.Metadata
			err = yaml.Unmarshal(b, &unmarshalledMetadata)
			Expect(err).NotTo(HaveOccurred())

			Expect(unmarshalledMetadata).To(Equal(inputMetadata))
		})
	})
})
