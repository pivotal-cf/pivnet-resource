package filesystem

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"gopkg.in/yaml.v2"
)

//go:generate counterfeiter . FileWriter

type FileWriter interface {
	WriteMetadataJSONFile(mdata metadata.Metadata) error
	WriteMetadataYAMLFile(mdata metadata.Metadata) error
	WriteVersionFile(versionWithETag string) error
}

type fileWriter struct {
	downloadDir string
	logger      *log.Logger
}

func NewFileWriter(downloadDir string, logger *log.Logger) FileWriter {
	return &fileWriter{
		downloadDir: downloadDir,
		logger:      logger,
	}
}

func (w fileWriter) WriteMetadataYAMLFile(mdata metadata.Metadata) error {
	yamlMetadataFilepath := filepath.Join(w.downloadDir, "metadata.yaml")
	w.logger.Println("Writing metadata to json file")

	yamlMetadata, err := yaml.Marshal(mdata)
	if err != nil {
		// Untested as it is too hard to force yaml.Marshal to return an error
		return err
	}

	err = ioutil.WriteFile(yamlMetadataFilepath, yamlMetadata, os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}

func (w fileWriter) WriteMetadataJSONFile(mdata metadata.Metadata) error {
	jsonMetadataFilepath := filepath.Join(w.downloadDir, "metadata.json")
	w.logger.Println("Writing metadata to json file")

	jsonMetadata, err := json.Marshal(mdata)
	if err != nil {
		// Untested as it is too hard to force json.Marshal to return an error
		return err
	}

	err = ioutil.WriteFile(jsonMetadataFilepath, jsonMetadata, os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}

func (w fileWriter) WriteVersionFile(versionWithETag string) error {
	versionFilepath := filepath.Join(w.downloadDir, "version")

	w.logger.Println("Writing version to file")

	err := ioutil.WriteFile(versionFilepath, []byte(versionWithETag), os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}
