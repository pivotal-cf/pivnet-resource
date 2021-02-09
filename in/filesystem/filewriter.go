package filesystem

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
	"gopkg.in/yaml.v2"
)

type FileWriter struct {
	downloadDir string
	logger      logger.Logger
}

func NewFileWriter(downloadDir string, logger logger.Logger) *FileWriter {
	return &FileWriter{
		downloadDir: downloadDir,
		logger:      logger,
	}
}

func (w FileWriter) WriteMetadataYAMLFile(mdata metadata.Metadata) error {
	yamlMetadataFilepath := filepath.Join(w.downloadDir, "metadata.yaml")
	w.logger.Debug("Writing metadata to yaml file")

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

func (w FileWriter) WriteMetadataJSONFile(mdata metadata.Metadata) error {
	jsonMetadataFilepath := filepath.Join(w.downloadDir, "metadata.json")
	w.logger.Debug("Writing metadata to json file")

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

func (w FileWriter) WriteVersionFile(version string) error {
	versionFilepath := filepath.Join(w.downloadDir, "version")

	w.logger.Debug("Writing version to file")

	err := ioutil.WriteFile(versionFilepath, []byte(version), os.ModePerm)
	if err != nil {
		// Untested as it is too hard to force io.WriteFile to return an error
		return err
	}

	return nil
}
