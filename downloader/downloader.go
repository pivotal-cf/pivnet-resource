package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/go-pivnet/v6/download"
)

//go:generate counterfeiter --fake-name FakeClient . client
type client interface {
	DownloadProductFile(writer *download.FileInfo, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
}

type Downloader struct {
	client         client
	downloadDir    string
	logger         logger.Logger
	progressWriter io.Writer
}

func NewDownloader(
	client client,
	downloadDir string,
	logger logger.Logger,
	progressWriter io.Writer,
) *Downloader {
	return &Downloader{
		client:         client,
		downloadDir:    downloadDir,
		logger:         logger,
		progressWriter: progressWriter,
	}
}

func (d Downloader) Download(
	pfs []pivnet.ProductFile,
	productSlug string,
	releaseID int,
) ([]string, error) {
	d.logger.Debug("Ensuring download directory exists")

	err := os.MkdirAll(d.downloadDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, pf := range pfs {
		parts := strings.Split(pf.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		downloadPath := filepath.Join(d.downloadDir, fileName)

		d.logger.Debug(fmt.Sprintf("Creating file: '%s'", downloadPath))
		file, err := os.Create(downloadPath)
		if err != nil {
			return nil, err
		}

		fileInfo, err := download.NewFileInfo(file)
		if err != nil {
			return nil, err
		}

		d.logger.Info(fmt.Sprintf(
			"Downloading: '%s' to file: '%s'",
			pf.Name,
			downloadPath,
		))

		err = d.client.DownloadProductFile(fileInfo, productSlug, releaseID, pf.ID, d.progressWriter)
		if err != nil {
			d.logger.Info(fmt.Sprintf("Download failed: %s",
				err.Error(),
			))
			return nil, err
		}
		fileNames = append(fileNames, downloadPath)
	}

	return fileNames, nil
}
