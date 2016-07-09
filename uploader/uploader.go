package uploader

import (
	"fmt"
	"path/filepath"
)

type Client interface {
	UploadFile(string) (string, error)
}

type client struct {
	filepathPrefix string
	sourcesDir     string

	transport Transport
}

type Config struct {
	FilepathPrefix string
	SourcesDir     string

	Transport Transport
}

func NewClient(config Config) Client {
	return &client{
		filepathPrefix: config.FilepathPrefix,
		sourcesDir:     config.SourcesDir,

		transport: config.Transport,
	}
}

func (c client) UploadFile(exactGlob string) (string, error) {
	if exactGlob == "" {
		return "", fmt.Errorf("glob must not be empty")
	}

	filename := filepath.Base(exactGlob)

	remoteDir := "product_files/" + c.filepathPrefix + "/"
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
