package downloader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

//go:generate counterfeiter --fake-name FakeExtendedClient . extendedClient
type extendedClient interface {
	DownloadFile(writer io.Writer, downloadLink string) error
}

type Downloader struct {
	extendedClient extendedClient
	downloadDir    string
	logger         *log.Logger
}

func NewDownloader(
	extendedClient extendedClient,
	downloadDir string,
	logger *log.Logger,
) *Downloader {
	return &Downloader{
		extendedClient: extendedClient,
		downloadDir:    downloadDir,
		logger:         logger,
	}
}

func (d Downloader) Download(downloadLinks map[string]string) ([]string, error) {
	d.logger.Println("Ensuring download directory exists")

	err := os.MkdirAll(d.downloadDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for fileName, downloadLink := range downloadLinks {
		downloadPath := filepath.Join(d.downloadDir, fileName)

		d.logger.Println(fmt.Sprintf("Creating file: '%s'", downloadPath))
		file, err := os.Create(downloadPath)
		if err != nil {
			return nil, err
		}

		d.logger.Println(fmt.Sprintf("Downloading link: '%s' to file: '%s'", downloadLink, downloadPath))
		err = d.extendedClient.DownloadFile(file, downloadLink)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, downloadPath)
	}

	return fileNames, nil
}
