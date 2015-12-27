package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/s3"
	"github.com/pivotal-cf-experimental/pivnet-resource/uploader"
)

const (
	s3OutBinaryName = "s3-out"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf("usage: %s <sources directory>", os.Args[0]))
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

	mustBeNonEmpty(input.Source.APIToken, "api_token")
	mustBeNonEmpty(input.Source.ProductName, "product_name")
	mustBeNonEmpty(input.Params.VersionFile, "version_file")
	mustBeNonEmpty(input.Params.ReleaseTypeFile, "release_type_file")
	mustBeNonEmpty(input.Params.EulaSlugFile, "eula_slug_file")

	if input.Params.FileGlob == "" && input.Params.FilepathPrefix == "" {
		fmt.Fprintln(os.Stderr, "file glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		mustBeNonEmpty(input.Source.AccessKeyID, "access_key_id")
		mustBeNonEmpty(input.Source.SecretAccessKey, "secret_access_key")
		mustBeNonEmpty(input.Params.FileGlob, "file glob")
		mustBeNonEmpty(input.Params.FilepathPrefix, "s3_filepath_prefix")

		s3Client := s3.NewClient(s3.NewClientConfig{
			AccessKeyID:     input.Source.AccessKeyID,
			SecretAccessKey: input.Source.SecretAccessKey,
			RegionName:      "eu-west-1",
			Bucket:          "pivotalnetwork",

			Stdout: os.Stdout,
			Stderr: os.Stderr,

			OutBinaryPath: filepath.Join(myDir, s3OutBinaryName),
		})
		uploaderClient := uploader.NewClient(uploader.Config{
			FileGlob:       input.Params.FileGlob,
			FilepathPrefix: input.Params.FilepathPrefix,
			SourcesDir:     sourcesDir,

			Transport: s3Client,
		})

		err := uploaderClient.Upload()

		if err != nil {
			log.Fatal(err)
		}
	}

	pivnetClient := pivnet.NewClient(pivnet.URL, input.Source.APIToken)

	config := pivnet.CreateReleaseConfig{
		ProductName:    input.Source.ProductName,
		ReleaseType:    readStringContents(sourcesDir, input.Params.ReleaseTypeFile),
		EulaSlug:       readStringContents(sourcesDir, input.Params.EulaSlugFile),
		ProductVersion: readStringContents(sourcesDir, input.Params.VersionFile),
		Description:    readStringContents(sourcesDir, input.Params.DescriptionFile),
		ReleaseDate:    readStringContents(sourcesDir, input.Params.ReleaseDateFile),
	}

	pivnetClient.CreateRelease(config)

	out := concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: config.ProductVersion,
		},
		Metadata: []string{},
	}

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
