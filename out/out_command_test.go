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
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
	"github.com/robdimsdale/sanitizer"
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
		etag                 string
		versionWithETag      string
		productID            int
		releaseID            int

		releaseType string
		eulaSlug    string

		releaseTypes []string
		eulas        []pivnet.EULA

		releaseTypesResponse       pivnet.ReleaseTypesResponse
		releaseTypesResponseStatus int

		eulasResponse       pivnet.EULAsResponse
		eulasResponseStatus int

		existingETags               []http.Header
		existingETagsResponseStatus int

		existingReleasesResponse       pivnet.ReleasesResponse
		existingReleasesResponseStatus int
		newReleaseResponse             pivnet.CreateReleaseResponse
		productsResponse               pivnet.Product

		newProductFileRequest            createProductFileBody
		newProductFileResponseStatusCode int
		newProductFileResponse           pivnet.ProductFile

		outRequest concourse.OutRequest
		outCommand *out.OutCommand
	)

	BeforeEach(func() {
		metadataFile = "metadata"
		metadataFilePath = ""
		metadataFileContents = ""

		server = ghttp.NewServer()

		releaseTypes = []string{
			"foo release",
			"bar",
			"third release type",
		}

		releaseTypesResponse = pivnet.ReleaseTypesResponse{
			ReleaseTypes: releaseTypes,
		}
		releaseTypesResponseStatus = http.StatusOK

		eulas = []pivnet.EULA{
			{
				ID:   1,
				Slug: "eulaSlug1",
			},
			{
				ID:   2,
				Slug: "eulaSlug2",
			},
		}

		eulasResponse = pivnet.EULAsResponse{
			EULAs: eulas,
		}
		eulasResponseStatus = http.StatusOK

		version = "2.1.3"
		etag = "etag-0"

		var err error
		versionWithETag, err = versions.CombineVersionAndETag(version, etag)
		Expect(err).NotTo(HaveOccurred())

		productID = 1
		releaseID = 2

		existingReleasesResponse = pivnet.ReleasesResponse{
			Releases: []pivnet.Release{
				{
					ID:      1234,
					Version: "some-other-version",
				},
			},
		}
		existingReleasesResponseStatus = http.StatusOK

		existingETags = []http.Header{
			{"ETag": []string{`"etag-0"`}},
		}
		existingETagsResponseStatus = http.StatusOK

		newReleaseResponse = pivnet.CreateReleaseResponse{
			Release: pivnet.Release{
				ID:      releaseID,
				Version: version,
				EULA: &pivnet.EULA{
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
			FileVersion:  version,
			MD5:          "220c7810f41695d9a87d70b68ccf2aeb", // hard-coded for now
			AWSObjectKey: fmt.Sprintf("product_files/Some-Case-Sensitive-Path/%s", productFileName0),
		}}

		productsResponse = pivnet.Product{
			ID:   productID,
			Slug: productSlug,
		}

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

sleep 0.1
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
		err = ioutil.WriteFile(versionFilePath, []byte(versionWithETag), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		releaseType = releaseTypes[0]
		releaseTypeFile = "release_type"
		releaseTypeFilePath := filepath.Join(sourcesDir, releaseTypeFile)
		err = ioutil.WriteFile(releaseTypeFilePath, []byte(releaseType), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		eulaSlug = eulas[0].Slug
		eulaSlugFile = "eula_slug"
		eulaSlugFilePath := filepath.Join(sourcesDir, eulaSlugFile)
		err = ioutil.WriteFile(eulaSlugFilePath, []byte(eulaSlug), os.ModePerm)
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
		metadataFilePath = filepath.Join(sourcesDir, metadataFile)
		err := ioutil.WriteFile(metadataFilePath, []byte(metadataFileContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/eulas", apiPrefix)),
				ghttp.RespondWithJSONEncoded(eulasResponseStatus, eulasResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/releases/release_types", apiPrefix)),
				ghttp.RespondWithJSONEncoded(releaseTypesResponseStatus, releaseTypesResponse),
			),
		)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.RespondWithJSONEncoded(
					existingReleasesResponseStatus,
					existingReleasesResponse,
				),
			),
		)

		for i, r := range existingReleasesResponse.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.RespondWith(existingETagsResponseStatus, nil, existingETags[i]),
				),
			)
		}

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

		etag := `"etag-0"`
		etagHeader := http.Header{"ETag": []string{etag}}
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, releaseID),
				),
				ghttp.RespondWith(http.StatusOK, nil, etagHeader),
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
				EULASlugFile:    eulaSlugFile,
				FilepathPrefix:  s3FilepathPrefix,
				MetadataFile:    metadataFile,
			},
		}

		sanitized := concourse.SanitizedSource(outRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
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
		response, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response.Version.ProductVersion).To(Equal(versionWithETag))
	})

	Context("when outDir is empty", func() {
		BeforeEach(func() {
			outDir = ""
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*out dir.*provided"))
			Expect(server.ReceivedRequests()).To(BeEmpty())
		})
	})

	Context("when getting existing releases returns error", func() {
		BeforeEach(func() {
			existingReleasesResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when getting existing releases etag returns error", func() {
		BeforeEach(func() {
			existingETagsResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when the s3-out exits with error", func() {
		BeforeEach(func() {
			s3OutScriptContents := `#!/bin/sh

sleep 0.1
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
			existingReleasesResponse = pivnet.ReleasesResponse{
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
			eulaSlugFile = ""
			releaseTypeFile = ""
			versionFile = ""

			version = "1.1.2"
			metadataFile = "metadata"
			metadataFilePath = filepath.Join(sourcesDir, metadataFile)
		})

		JustBeforeEach(func() {
			err := ioutil.WriteFile(metadataFilePath, []byte(metadataFileContents), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when it has a release key", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
           release:
             eula_slug: %s
             version: "1.1.2#etag-0"
             release_type: %s`,
					eulas[0].Slug,
					releaseTypes[0],
				)
			})

			It("overrides any other files specifying metadata", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when a duplicate version exists", func() {
				BeforeEach(func() {
					existingReleasesResponse = pivnet.ReleasesResponse{
						Releases: []pivnet.Release{
							{
								ID:      1234,
								Version: "1.1.2",
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

				Expect(err.Error()).To(MatchRegexp(".*metadata_file.*could not be parsed"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		Context("when metadata file contains matching product file descriptions", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
           release:
            eula_slug: %s
            version: "1.1.2#etag-0"
            release_type: %s
           product_files:
           - file: %s
             description: |
               some
               multi-line
               description`,
					eulas[0].Slug,
					releaseTypes[0],
					productFileRelativePath0,
				)

				newProductFileRequest.ProductFile.Description = "some\nmulti-line\ndescription"
			})

			It("creates product files with the matching description", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when metadata file contains matching product file without descriptions", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
          release:
           eula_slug: %s
           version: "1.1.2#etag-0"
           release_type: %s
          product_files:
          - file: %s`,
					eulas[0].Slug,
					releaseTypes[0],
					productFileRelativePath0,
				)
			})

			It("creates product files with the matching description", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when metadata file contains a file that does not correspond to any glob-matched file", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
           release:
            eula_slug: %s
            version: "1.1.2#etag-0"
            release_type: %s
           product_files:
           - file: not-a-real-file
             description: |
               some
               multi-line
               description
           - file: also-not-a-real-file
             description: |
               some
               other
               description`,
					eulas[0].Slug,
					releaseTypes[0],
				)
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*metadata.*not-a-real-file.*also-not-a-real-file"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		Context("when metadata file contains an empty value for file", func() {
			BeforeEach(func() {
				metadataFileContents =
					`---
           product_files:
           - file: `
			})

			It("returns an error", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*metadata.*empty.*file"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		Context("when metadata file contains upload_as for valid file", func() {
			BeforeEach(func() {
				metadataFileContents = fmt.Sprintf(
					`---
          release:
           eula_slug: %s
           version: "1.1.2#etag-0"
           release_type: %s
          product_files:
          - file: %s
            upload_as: some_remote_file`,
					eulas[0].Slug,
					releaseTypes[0],
					productFileRelativePath0,
				)

				newProductFileRequest.ProductFile.Name = "some_remote_file"
			})

			It("creates product files with the provided name", func() {
				_, err := outCommand.Run(outRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when globbing fails", func() {
		BeforeEach(func() {
			fileGlob = "}{"
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*no matches.*}{"))
			Expect(server.ReceivedRequests()).To(BeEmpty())
		})
	})

	Context("when there is an error getting release types", func() {
		BeforeEach(func() {
			releaseTypesResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when the release type is invalid", func() {
		BeforeEach(func() {
			// Use metadata rather than release type file
			releaseTypeFile = ""
			releaseType = "not a valid release type"
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*release_type.*one of"))
			Expect(err.Error()).To(ContainSubstring(releaseTypes[0]))
			Expect(err.Error()).To(ContainSubstring(releaseTypes[1]))
			Expect(err.Error()).To(ContainSubstring(releaseTypes[2]))
		})
	})

	Context("when there is an error getting eulas", func() {
		BeforeEach(func() {
			eulasResponseStatus = http.StatusNotFound
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("404"))
		})
	})

	Context("when the release type is invalid", func() {
		BeforeEach(func() {
			// Use metadata rather than release type file
			eulaSlugFile = ""
			eulaSlug = "not a valid eula"
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*eula.*one of"))
			Expect(err.Error()).To(ContainSubstring(eulas[0].Slug))
			Expect(err.Error()).To(ContainSubstring(eulas[1].Slug))
		})
	})
})

type createProductFileBody struct {
	ProductFile pivnet.ProductFile `json:"product_file"`
}
