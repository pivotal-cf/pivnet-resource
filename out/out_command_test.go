package out_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/out"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
)

var _ = Describe("Out", func() {
	var (
		tempDir string

		filesToUploadDirName string

		uploadFilesSourceDir     string
		productFileName0         string
		productFileFullPath0     string
		productFileRelativePath0 string

		server *ghttp.Server

		ginkgoLogger logger.Logger

		productSlug string

		accessKeyID     string
		secretAccessKey string
		apiToken        string

		outDir          string
		sourcesDir      string
		logFilePath     string
		s3OutBinaryName string

		fileGlob         string
		versionFile      string
		releaseTypeFile  string
		eulaSlugFile     string
		s3FilepathPrefix string
		metadataFile     string
		metadataFilePath string

		metadataFileContents string
		version              string
		productID            int
		releaseID            int

		existingReleasesResponse pivnet.Response
		newReleaseResponse       pivnet.CreateReleaseResponse
		productsResponse         pivnet.Product

		newProductFileRequest            createProductFileBody
		newProductFileResponseStatusCode int
		newProductFileResponse           pivnet.ProductFile

		outRequest concourse.OutRequest
		outCommand *out.OutCommand
	)

	BeforeEach(func() {
		metadataFile = ""
		metadataFilePath = ""
		metadataFileContents = ""

		server = ghttp.NewServer()

		version = "2.1.3"

		productID = 1
		releaseID = 2

		existingReleasesResponse = pivnet.Response{
			Releases: []pivnet.Release{
				{
					ID:      1234,
					Version: "some-other-version",
				},
			},
		}

		newReleaseResponse = pivnet.CreateReleaseResponse{
			Release: pivnet.Release{
				ID: releaseID,
				Eula: &pivnet.Eula{
					Slug: "some-eula",
				},
			},
		}

		productSlug = "some-product-name"
		productFileName0 = "some-file"

		newProductFileResponseStatusCode = http.StatusCreated
		newProductFileRequest = createProductFileBody{pivnet.ProductFile{
			FileType:     "Software",
			Name:         productFileName0,
			MD5:          "220c7810f41695d9a87d70b68ccf2aeb", // hard-coded for now
			AWSObjectKey: fmt.Sprintf("product_files/Some-Case-Sensitive-Path/%s", productFileName0),
		}}

		productsResponse = pivnet.Product{
			ID:   productID,
			Slug: productSlug,
		}

		var err error
		outDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		sourcesDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "pivnet-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		s3OutBinaryName = "s3-out"
		s3OutScriptContents := `#!/bin/sh

echo "$@"`

		s3OutBinaryPath := filepath.Join(outDir, s3OutBinaryName)
		err = ioutil.WriteFile(s3OutBinaryPath, []byte(s3OutScriptContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		apiToken = "some-api-token"
		accessKeyID = "some-access-key-id"
		secretAccessKey = "some-secret-access-key"

		filesToUploadDirName = "files_to_upload"

		fileGlob = fmt.Sprintf("%s/*", filesToUploadDirName)
		s3FilepathPrefix = "Some-Case-Sensitive-Path"

		versionFile = "version"
		versionFilePath := filepath.Join(sourcesDir, versionFile)
		err = ioutil.WriteFile(versionFilePath, []byte(version), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		releaseTypeFile = "release_type"
		releaseTypeFilePath := filepath.Join(sourcesDir, releaseTypeFile)
		err = ioutil.WriteFile(releaseTypeFilePath, []byte("some_release"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		eulaSlugFile = "eula_slug"
		eulaSlugFilePath := filepath.Join(sourcesDir, eulaSlugFile)
		err = ioutil.WriteFile(eulaSlugFilePath, []byte("some_eula"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		uploadFilesSourceDir = filepath.Join(sourcesDir, filesToUploadDirName)
		err = os.Mkdir(uploadFilesSourceDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		productFileFullPath0 = filepath.Join(uploadFilesSourceDir, productFileName0)
		productFileRelativePath0 = filepath.Join(filesToUploadDirName, productFileName0)
		err = ioutil.WriteFile(productFileFullPath0, []byte("some contents"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(outDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(sourcesDir)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncoded(http.StatusOK, existingReleasesResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncoded(http.StatusCreated, newReleaseResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncoded(http.StatusOK, productsResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf("%s/products/%s/product_files", apiPrefix, productSlug),
				),
				ghttp.VerifyJSONRepresenting(newProductFileRequest),
				ghttp.RespondWithJSONEncoded(newProductFileResponseStatusCode, newProductFileResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"PATCH",
					fmt.Sprintf(
						"%s/products/%d/releases/%d/add_product_file",
						apiPrefix,
						productID,
						releaseID,
					),
				),
				ghttp.RespondWith(http.StatusNoContent, ""),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"PATCH",
					fmt.Sprintf(
						"%s/products/%s/releases/%d",
						apiPrefix,
						productSlug,
						releaseID,
					),
				),
				ghttp.RespondWithJSONEncoded(http.StatusOK, newReleaseResponse),
			),
		)

		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        apiToken,
				ProductSlug:     productSlug,
				Endpoint:        server.URL(),
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
			Params: concourse.OutParams{
				FileGlob:        fileGlob,
				VersionFile:     versionFile,
				ReleaseTypeFile: releaseTypeFile,
				EulaSlugFile:    eulaSlugFile,
				FilepathPrefix:  s3FilepathPrefix,
				MetadataFile:    metadataFile,
			},
		}

		sanitized := concourse.SanitizedSource(outRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2"
		outCommand = out.NewOutCommand(out.OutCommandConfig{
			BinaryVersion:   binaryVersion,
			Logger:          ginkgoLogger,
			OutDir:          outDir,
			SourcesDir:      sourcesDir,
			LogFilePath:     logFilePath,
			S3OutBinaryName: s3OutBinaryName,
		})
	})

	It("runs without error", func() {
		_, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("input validation", func() {
		Context("when outDir is empty", func() {
			BeforeEach(func() {
				outDir = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*out dir.*provided"))
			})
		})

		Context("when no api token is provided", func() {
			BeforeEach(func() {
				apiToken = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
			})
		})

		Context("when no product slug is provided", func() {
			BeforeEach(func() {
				productSlug = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
			})
		})

		Context("when no aws access key id is provided", func() {
			BeforeEach(func() {
				accessKeyID = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*access_key_id.*provided"))
			})
		})

		Context("when no aws secret access key is provided", func() {
			BeforeEach(func() {
				secretAccessKey = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*secret_access_key.*provided"))
			})
		})

		Context("when file glob is not provided", func() {
			BeforeEach(func() {
				fileGlob = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*file glob.*provided"))
			})
		})

		Context("when s3 filepath prefix is not provided", func() {
			BeforeEach(func() {
				s3FilepathPrefix = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*s3_filepath_prefix.*provided"))
			})
		})

		Context("when version file is not provided", func() {
			BeforeEach(func() {
				versionFile = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*version_file.*provided"))
			})
		})

		Context("when release_type file is not provided", func() {
			BeforeEach(func() {
				releaseTypeFile = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*release_type_file.*provided"))
			})
		})

		Context("when eula_slug file is not provided", func() {
			BeforeEach(func() {
				eulaSlugFile = ""
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*eula_slug_file.*provided"))
			})
		})

		Context("when metadata file is provided", func() {
			BeforeEach(func() {
				metadataFile = "metadata"
			})

			It("returns an error (metadata file does not exist)", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*metadata_file.*could not be read"))
			})

			Context("when metadata file exists", func() {
				BeforeEach(func() {
					metadataFileContents = ``
					metadataFilePath = filepath.Join(sourcesDir, metadataFile)

					err := ioutil.WriteFile(metadataFilePath, []byte(metadataFileContents), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				Context("when metadata file contains invalid yaml", func() {
					BeforeEach(func() {
						metadataFileContents = "{{"

						err := ioutil.WriteFile(metadataFilePath, []byte(metadataFileContents), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := outCommand.Run(outRequest)
						Expect(err).To(HaveOccurred())

						Expect(err.Error()).To(MatchRegexp(".*metadata_file.*invalid"))
					})
				})
			})
		})
	})

	Context("when the s3-out exits with error", func() {
		BeforeEach(func() {
			s3OutScriptContents := `#!/bin/sh

sleep 1
exit 1`

			s3OutBinaryPath := filepath.Join(outDir, s3OutBinaryName)
			err := ioutil.WriteFile(s3OutBinaryPath, []byte(s3OutScriptContents), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*running.*%s.*", s3OutBinaryName))
		})
	})

	Context("when a release already exists with the expected version", func() {
		BeforeEach(func() {
			existingReleasesResponse = pivnet.Response{
				Releases: []pivnet.Release{
					{
						ID:      1234,
						Version: version,
					},
				},
			}
		})

		It("exits with error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*release already exists.*%s.*", version))
		})
	})

	Context("when creating a new product file fails", func() {
		BeforeEach(func() {
			newProductFileResponseStatusCode = http.StatusForbidden
		})

		It("exits with error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*returned.*403.*"))
		})
	})

	Context("when metadata file is provided", func() {
		BeforeEach(func() {
			metadataFile = "metadata"
			metadataFilePath = filepath.Join(sourcesDir, metadataFile)
		})

		JustBeforeEach(func() {
			err := ioutil.WriteFile(metadataFilePath, []byte(metadataFileContents), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when metadata file contains matching product file descriptions", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
           product_files:
           - file: %s
             description: |
               some
               multi-line
               description`,
					productFileRelativePath0,
				)

				newProductFileRequest.ProductFile.Description = "some\nmulti-line\ndescription"
			})

			It("creates product files with the matching description", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

type createProductFileBody struct {
	ProductFile pivnet.ProductFile `json:"product_file"`
}
