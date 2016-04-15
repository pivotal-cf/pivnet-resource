package in_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/in"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

var _ = Describe("In", func() {
	var (
		server *ghttp.Server

		releaseID int

		file1URLPath         string
		productFiles         []pivnet.ProductFile
		productFilesResponse pivnet.ProductFiles

		downloadDir string

		ginkgoLogger logger.Logger

		productVersion  string
		etag            string
		versionWithETag string

		inRequest              concourse.InRequest
		inCommand              *in.InCommand
		pivnetReleasesResponse *pivnet.ReleasesResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		productVersion = "C"
		etag = "etag-0"

		var err error
		versionWithETag, err = versions.CombineVersionAndETag(productVersion, etag)
		Expect(err).NotTo(HaveOccurred())

		releaseID = 1234
		file1URLPath = "/file1"
		file1URL := fmt.Sprintf("%s%s", server.URL(), file1URLPath)
		productFiles = []pivnet.ProductFile{
			{
				ID:           1234,
				Name:         "product file 1234",
				Description:  "some product file 1234",
				AWSObjectKey: "some-key 1234",
				FileType:     "some-file-type 1234",
				FileVersion:  "some-file-version 1234",
				MD5:          "some-md5 1234",
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "foo",
					},
				},
			},
			{
				ID:           3456,
				Name:         "product file 3456",
				Description:  "some product file 3456",
				AWSObjectKey: "some-key 3456",
				FileType:     "some-file-type 3456",
				FileVersion:  "some-file-version 3456",
				MD5:          "some-md5 3456",
				Links: &pivnet.Links{
					Download: map[string]string{
						"href": "bar",
					},
				},
			},
		}
		productFilesResponse = pivnet.ProductFiles{
			ProductFiles: productFiles,
		}

		pivnetReleasesResponse = &pivnet.ReleasesResponse{
			Releases: []pivnet.Release{
				{
					Version: "A",
				},
				{
					Version: productVersion,
					ID:      releaseID,
					Links: &pivnet.Links{
						ProductFiles: map[string]string{
							"href": file1URL,
						},
					},
				},
				{
					Version: "B",
				},
			},
		}

		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				Endpoint:    server.URL(),
			},
			Version: concourse.Version{
				versionWithETag,
			},
		}

	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncoded(http.StatusOK, pivnetReleasesResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf(
						"%s/products/%s/releases/%d/eula_acceptance",
						apiPrefix,
						productSlug,
						releaseID,
					),
				),
				ghttp.RespondWith(http.StatusOK, ""),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					file1URLPath,
				),
				ghttp.RespondWithJSONEncoded(http.StatusOK, productFilesResponse),
			),
		)

		for _, p := range productFiles {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files/%d",
							apiPrefix,
							productSlug,
							releaseID,
							p.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, p),
				),
			)
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		inCommand = in.NewInCommand(binaryVersion, ginkgoLogger, downloadDir)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes a version file with the downloaded version and etag", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "version")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(versionContents)).To(Equal(versionWithETag))
	})

	var validateProductFilesMetadata = func(
		writtenMetadata metadata.Metadata,
		productFiles []pivnet.ProductFile,
	) {
		Expect(writtenMetadata.ProductFiles).To(HaveLen(len(productFiles)))
		for i, p := range productFiles {
			Expect(writtenMetadata.ProductFiles[i].File).To(Equal(p.Name))
			Expect(writtenMetadata.ProductFiles[i].Description).To(Equal(p.Description))
			Expect(writtenMetadata.ProductFiles[i].ID).To(Equal(p.ID))
			Expect(writtenMetadata.ProductFiles[i].AWSObjectKey).To(Equal(p.AWSObjectKey))
			Expect(writtenMetadata.ProductFiles[i].FileType).To(Equal(p.FileType))
			Expect(writtenMetadata.ProductFiles[i].FileVersion).To(Equal(p.FileVersion))
			Expect(writtenMetadata.ProductFiles[i].MD5).To(Equal(p.MD5))
			Expect(writtenMetadata.ProductFiles[i].UploadAs).To(BeEmpty())
		}
	}

	It("writes a metadata file in yaml format", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "metadata.yaml")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())

		var writtenMetadata metadata.Metadata
		err = yaml.Unmarshal(versionContents, &writtenMetadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(writtenMetadata.Release).NotTo(BeNil())
		Expect(writtenMetadata.Release.Version).To(Equal(productVersion))

		validateProductFilesMetadata(writtenMetadata, productFiles)
	})

	It("writes a metadata file in json format", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		versionFilepath := filepath.Join(downloadDir, "metadata.json")
		versionContents, err := ioutil.ReadFile(versionFilepath)
		Expect(err).NotTo(HaveOccurred())

		var writtenMetadata metadata.Metadata
		err = json.Unmarshal(versionContents, &writtenMetadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(writtenMetadata.Release).NotTo(BeNil())
		Expect(writtenMetadata.Release.Version).To(Equal(productVersion))

		validateProductFilesMetadata(writtenMetadata, productFiles)
	})

	It("does not download any of the files in the specified release", func() {
		_, err := inCommand.Run(inRequest)
		Expect(err).NotTo(HaveOccurred())

		files, err := ioutil.ReadDir(downloadDir)
		Expect(err).ShouldNot(HaveOccurred())

		// the version and metadata files will always exist
		Expect(len(files)).To(Equal(3))
		Expect(files[0].Name()).To(Equal("metadata.json"))
		Expect(files[1].Name()).To(Equal("metadata.yaml"))
		Expect(files[2].Name()).To(Equal("version"))
	})

	Context("when version is provided without etag", func() {
		BeforeEach(func() {
			inRequest.Version = concourse.Version{
				ProductVersion: productVersion,
			}
		})

		It("returns without error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when release has no links", func() {
		BeforeEach(func() {
			pivnetReleasesResponse.Releases[1].Links = nil
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp("Failed to get Product File"))
		})
	})
})
