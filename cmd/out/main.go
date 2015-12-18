package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
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

	if input.Params.File == "" && input.Params.FilepathPrefix == "" {
		fmt.Fprintln(os.Stderr, "file glob and s3_filepath_prefix not provided - skipping upload to s3")
	} else {
		if input.Source.AccessKeyID == "" {
			log.Fatalln("access_key_id must be provided")
		}

		if input.Source.SecretAccessKey == "" {
			log.Fatalln("secret_access_key must be provided")
		}

		if input.Params.File == "" {
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
			input.Params.File,
			"product_files/"+input.Params.FilepathPrefix+"/",
			sourcesDir,
		)

		if err != nil {
			log.Fatal(err)
		}
	}
}
