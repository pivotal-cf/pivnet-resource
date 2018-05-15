package uploader

import (
	"fmt"
	"path/filepath"
	"strings"
)

//go:generate counterfeiter --fake-name FakeTransport . transport
type transport interface {
	Upload(fileGlob string, filepathPrefix string, sourcesDir string) error
}

type Client struct {
	filepathPrefix string
	sourcesDir     string

	transport transport
}

type Config struct {
	FilepathPrefix string
	SourcesDir     string

	Transport transport
}

func NewClient(config Config) *Client {
	return &Client{
		filepathPrefix: config.FilepathPrefix,
		sourcesDir:     config.SourcesDir,

		transport: config.Transport,
	}
}

func (c Client) UploadFile(exactGlob string) (error) {
	
	_, remoteDir, err := c.ComputeAWSObjectKey(exactGlob)
	if err != nil {
		return err
	}

	err = c.transport.Upload(
		exactGlob,
		remoteDir,
		c.sourcesDir,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) ComputeAWSObjectKey(exactGlob string) (string, string, error) {
	if exactGlob == "" {
		return "", "", fmt.Errorf("glob must not be empty")
	}

	filename := filepath.Base(exactGlob)

	var remoteDir string
	switch {
	case strings.HasPrefix(c.filepathPrefix, "product-files"):
		remoteDir = c.filepathPrefix + "/"
	case strings.HasPrefix(c.filepathPrefix, "product_files"):
		remoteDir = c.filepathPrefix + "/"
	default:
		remoteDir = "product_files/" + c.filepathPrefix + "/"
	}

	remotePath := fmt.Sprintf("%s%s", remoteDir, filename)
	return remotePath, remoteDir, nil

}