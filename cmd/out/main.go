package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
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

	s3Input := S3Request{
		Source: S3Source{
			AccessKeyID:     input.Source.AccessKeyID,
			SecretAccessKey: input.Source.SecretAccessKey,
			RegionName:      "eu-west-1",
			Bucket:          "pivotalnetwork",
		},
		Params: S3Params{
			File: input.Params.File,
			To:   "product_files/" + input.Params.FilepathPrefix + "/",
		},
	}

	s3OutPath := filepath.Join(myDir, s3OutBinaryName)

	cmd := exec.Command(s3OutPath, sourcesDir)

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln(err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewEncoder(cmdIn).Encode(s3Input)
	if err != nil {
		log.Fatalln(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

type S3Request struct {
	Params S3Params `json:"params"`
	Source S3Source `json:"source"`
}

type S3Params struct {
	File string `json:"file"`
	To   string `json:"to"`
}

type S3Source struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket"`
	RegionName      string `json:"region_name"`
	Regexp          string `json:"regexp"`
}
