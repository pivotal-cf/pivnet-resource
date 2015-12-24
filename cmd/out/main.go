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

	if input.Source.APIToken == "" {
		log.Fatalln("api_token must be provided")
	}

	if input.Source.ProductName == "" {
		log.Fatalln("product_name must be provided")
	}

	if input.Params.VersionFile == "" {
		log.Fatalln("version_file must be provided")
	}

	if input.Params.ReleaseTypeFile == "" {
		log.Fatalln("release_type_file must be provided")
	}

	if input.Params.EulaSlugFile == "" {
		log.Fatalln("eula_slug_file must be provided")
	}

	if input.Params.FileGlob == "" && input.Params.FilepathPrefix == "" {
		fmt.Fprintln(os.Stderr, "file glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		if input.Source.AccessKeyID == "" {
			log.Fatalln("access_key_id must be provided")
		}

		if input.Source.SecretAccessKey == "" {
			log.Fatalln("secret_access_key must be provided")
		}

		if input.Params.FileGlob == "" {
			log.Fatalln("file glob must be provided")
		}

		if input.Params.FilepathPrefix == "" {
			log.Fatalln("s3_filepath_prefix must be provided")
		}

		s3Client := s3.NewClient(s3.NewClientConfig{
			AccessKeyID:     input.Source.AccessKeyID,
			SecretAccessKey: input.Source.SecretAccessKey,
			RegionName:      "eu-west-1",
			Bucket:          "pivotalnetwork",

			Stdout: os.Stdout,
			Stderr: os.Stderr,

			OutBinaryPath: filepath.Join(myDir, s3OutBinaryName),
		})

		err = s3Client.Out(
			input.Params.FileGlob,
			"product_files/"+input.Params.FilepathPrefix+"/",
			sourcesDir,
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	pivnetClient := pivnet.NewClient(pivnet.URL, input.Source.APIToken)

	releaseTypeFilepath := filepath.Join(sourcesDir, input.Params.ReleaseTypeFile)
	releaseTypeContents, err := ioutil.ReadFile(releaseTypeFilepath)
	if err != nil {
		log.Fatal(err)
	}

	releaseType := string(releaseTypeContents)

	var releaseDate string
	if input.Params.ReleaseDateFile != "" {
		releaseDateFilepath := filepath.Join(sourcesDir, input.Params.ReleaseDateFile)
		releaseDateContents, err := ioutil.ReadFile(releaseDateFilepath)
		if err != nil {
			log.Fatal(err)
		}
		releaseDate = string(releaseDateContents)
	}

	eulaSlugFilepath := filepath.Join(sourcesDir, input.Params.EulaSlugFile)
	eulaSlugContents, err := ioutil.ReadFile(eulaSlugFilepath)
	if err != nil {
		log.Fatal(err)
	}
	eulaSlug := string(eulaSlugContents)

	versionFilepath := filepath.Join(sourcesDir, input.Params.VersionFile)
	versionContents, err := ioutil.ReadFile(versionFilepath)
	if err != nil {
		log.Fatal(err)
	}

	productVersion := string(versionContents)

	productName := input.Source.ProductName

	pivnetClient.CreateRelease(pivnet.CreateReleaseConfig{
		ProductName:    productName,
		ProductVersion: productVersion,
		ReleaseType:    releaseType,
		ReleaseDate:    releaseDate,
		EulaSlug:       eulaSlug,
	})

	out := concourse.OutResponse{
		Version: concourse.Version{
			ProductVersion: productVersion,
		},
		Metadata: []string{},
	}

	err = json.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		log.Fatalln(err)
	}
}
