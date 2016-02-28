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

		uploadFilesSourceDir string

		server *ghttp.Server

		ginkgoLogger logger.Logger

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

		outRequest concourse.OutRequest
		outCommand *out.OutCommand
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		productID := 1
		releaseID := 2

		newReleaseResponse := pivnet.CreateReleaseResponse{
			Release: pivnet.Release{
				ID: releaseID,
				Eula: &pivnet.Eula{
					Slug: "some-eula",
				},
			},
		}

		productsResponse := pivnet.Product{
			ID:   productID,
			Slug: productSlug,
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
				ghttp.RespondWith(http.StatusCreated, ""),
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

		s3OutBinaryPath := filepath.Join(outDir, s3OutBinaryName)
		err = ioutil.WriteFile(s3OutBinaryPath, []byte(`#!/bin/sh

echo "$@"`), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		apiToken = "some-api-token"
		accessKeyID = "some-access-key-id"
		secretAccessKey = "some-secret-access-key"

		filesToUploadDirName := "files_to_upload"

		fileGlob = fmt.Sprintf("%s/*", filesToUploadDirName)
		s3FilepathPrefix = "Some-Case-Sensitive-Path"

		versionFile = "version"
		versionFilePath := filepath.Join(sourcesDir, versionFile)
		err = ioutil.WriteFile(versionFilePath, []byte("2.1.3"), os.ModePerm)
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

		fileToUploadPath := filepath.Join(uploadFilesSourceDir, "file-to-upload")
		err = ioutil.WriteFile(fileToUploadPath, []byte("some contents"), os.ModePerm)
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
})
