package uploader

import (
	"fmt"
	"path/filepath"
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

func (c Client) UploadFile(exactGlob string) (string, error) {
	if exactGlob == "" {
		return "", fmt.Errorf("glob must not be empty")
	}

	filename := filepath.Base(exactGlob)

	remoteDir := c.filepathPrefix + "/"
	remotePath := fmt.Sprintf("%s%s", remoteDir, filename)

	err := c.transport.Upload(
		exactGlob,
		remoteDir,
		c.sourcesDir,
	)
	if err != nil {
		return "", err
	}

	return remotePath, nil
}
