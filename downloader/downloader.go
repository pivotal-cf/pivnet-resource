package downloader

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet"
)

//go:generate counterfeiter --fake-name FakeClient . client
type client interface {
	DownloadProductFile(writer io.Writer, productSlug string, releaseID int, productFileID int) error
}

type Downloader struct {
	client      client
	downloadDir string
	logger      *log.Logger
}

func NewDownloader(
	client client,
	downloadDir string,
	logger *log.Logger,
) *Downloader {
	return &Downloader{
		client:      client,
		downloadDir: downloadDir,
		logger:      logger,
	}
}

func (d Downloader) Download(
	pfs []pivnet.ProductFile,
	productSlug string,
	releaseID int,
) ([]string, error) {
	d.logger.Println("Ensuring download directory exists")

	err := os.MkdirAll(d.downloadDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, pf := range pfs {
		parts := strings.Split(pf.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		downloadPath := filepath.Join(d.downloadDir, fileName)

		d.logger.Printf("Creating file: '%s'", downloadPath)
		file, err := os.Create(downloadPath)
		if err != nil {
			return nil, err
		}

		d.logger.Printf(
			"Downloading: '%s' to file: '%s'",
			pf.Name,
			downloadPath,
		)
		err = d.client.DownloadProductFile(file, productSlug, releaseID, pf.ID)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, downloadPath)
	}

	return fileNames, nil
}
