package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/s3"
	"github.com/pivotal-cf-experimental/pivnet-resource/sanitizer"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
)

const (
	s3OutBinaryName = "s3-out"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf(
			"not enough args - usage: %s <sources directory>", os.Args[0]))
	}

	sourcesDir := os.Args[1]

	myDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	var input concourse.OutRequest

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := ioutil.TempFile("", "pivnet-resource-out.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "PivNet Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "logging to %s\n", logFile.Name())

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l := logger.NewLogger(sanitizer)

	mustBeNonEmpty(input.Source.APIToken, "api_token")
	mustBeNonEmpty(input.Source.ProductSlug, "product_slug")
	mustBeNonEmpty(input.Params.VersionFile, "version_file")
	mustBeNonEmpty(input.Params.ReleaseTypeFile, "release_type_file")
	mustBeNonEmpty(input.Params.EulaSlugFile, "eula_slug_file")

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	if !skipUpload {
		mustBeNonEmpty(input.Source.AccessKeyID, "access_key_id")
		mustBeNonEmpty(input.Source.SecretAccessKey, "secret_access_key")
		mustBeNonEmpty(input.Params.FileGlob, "file glob")
		mustBeNonEmpty(input.Params.FilepathPrefix, "s3_filepath_prefix")
	}

	l.Debugf("Received input: %+v\n", input)

	clientConfig := pivnet.NewClientConfig{
		URL:       pivnet.URL,
		Token:     input.Source.APIToken,
		UserAgent: fmt.Sprintf("pivnet-resource/%s", version),
	}
	pivnetClient := pivnet.NewClient(
		clientConfig,
		l,
	)

	productSlug := input.Source.ProductSlug

	config := pivnet.CreateReleaseConfig{
		ProductSlug:     productSlug,
		ReleaseType:     readStringContents(sourcesDir, input.Params.ReleaseTypeFile),
		EulaSlug:        readStringContents(sourcesDir, input.Params.EulaSlugFile),
		ProductVersion:  readStringContents(sourcesDir, input.Params.VersionFile),
		Description:     readStringContents(sourcesDir, input.Params.DescriptionFile),
		ReleaseNotesURL: readStringContents(sourcesDir, input.Params.ReleaseNotesURLFile),
		ReleaseDate:     readStringContents(sourcesDir, input.Params.ReleaseDateFile),
	}

	release, err := pivnetClient.CreateRelease(config)
	if err != nil {
		log.Fatalln(err)
	}

	if skipUpload {
		l.Debugf("File glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		s3Client := s3.NewClient(s3.NewClientConfig{
			AccessKeyID:     input.Source.AccessKeyID,
			SecretAccessKey: input.Source.SecretAccessKey,
			RegionName:      "eu-west-1",
			Bucket:          "pivotalnetwork",

			Logger: l,

			Stdout: os.Stdout,
			Stderr: logFile,

			OutBinaryPath: filepath.Join(myDir, s3OutBinaryName),
		})

		uploaderClient := uploader.NewClient(uploader.Config{
			FileGlob:       input.Params.FileGlob,
			FilepathPrefix: input.Params.FilepathPrefix,
			SourcesDir:     sourcesDir,

			Logger: l,

			Transport: s3Client,
		})

		files, err := uploaderClient.Upload()
		for filename, remotePath := range files {
			product, err := pivnetClient.FindProductForSlug(productSlug)
			if err != nil {
				log.Fatalln(err)
			}

			l.Debugf(
				"Creating product file: {product_slug: %s, filename: %s, aws_object_key: %s, file_version: %s}\n",
				productSlug,
				filename,
				remotePath,
				release.Version,
			)

			productFile, err := pivnetClient.CreateProductFile(pivnet.CreateProductFileConfig{
				ProductSlug:  productSlug,
				Name:         filename,
				AWSObjectKey: remotePath,
				FileVersion:  release.Version,
			})
			if err != nil {
				log.Fatalln(err)
			}

			l.Debugf(
				"Adding product file: {product_slug: %s, product_id: %d, filename: %s, product_file_id: %d}\n",
				productSlug,
				product.ID,
				filename,
				productFile.ID,
			)

			err = pivnetClient.AddProductFile(product.ID, release.ID, productFile.ID)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	availability := readStringContents(sourcesDir, input.Params.AvailabilityFile)
	if availability != "Admins Only" {
		releaseUpdate := pivnet.Release{
			ID:           release.ID,
			Availability: availability,
		}
		release, err = pivnetClient.UpdateRelease(productSlug, releaseUpdate)
		if err != nil {
			log.Fatalln(err)
		}
	}

	out := concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: release.Version,
		},
		Metadata: []concourse.Metadata{
			{Name: "release_type", Value: release.ReleaseType},
			{Name: "release_date", Value: release.ReleaseDate},
			{Name: "description", Value: release.Description},
			{Name: "release_notes_url", Value: release.ReleaseNotesURL},
			{Name: "eula_slug", Value: release.Eula.Slug},
			{Name: "availability", Value: release.Availability},
		},
	}

	l.Debugf("Returning output: %+v\n", out)

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}

func mustBeNonEmpty(input string, key string) {
	if input == "" {
		log.Fatalf("%s must be provided\n", key)
	}
}

func readStringContents(sourcesDir, file string) string {
	if file == "" {
		return ""
	}
	fullPath := filepath.Join(sourcesDir, file)
	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}
