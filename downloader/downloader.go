package downloader

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
)

//go:generate counterfeiter --fake-name FakeClient . client
type client interface {
	DownloadProductFile(writer io.Writer, productSlug string, releaseID int, productFileID int) error
}

type Downloader struct {
	client      client
	downloadDir string
	logger      logger.Logger
}

func NewDownloader(
	client client,
	downloadDir string,
	logger logger.Logger,
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

		d.logger.Info(fmt.Sprintf(
			"Downloading: '%s' to file: '%s'",
			pf.Name,
			downloadPath,
		))

		maxAttempts := 3
		err = d.downloadProductFileWithRetries(file, productSlug, releaseID, pf.ID, maxAttempts)
		if err != nil {
			d.logger.Info(fmt.Sprintf("Download failed after %d attempts: %s",
				maxAttempts,
				err.Error(),
			))
			return nil, err
		}
		fileNames = append(fileNames, downloadPath)
	}

	return fileNames, nil
}

func (d Downloader) downloadProductFileWithRetries(
	file io.Writer,
	productSlug string,
	releaseID int,
	productFileID int,
	maxAttempts int,
) error {
	var err error

	for i := 0; i < maxAttempts; i++ {
		err = d.client.DownloadProductFile(file, productSlug, releaseID, productFileID)

		if err != nil {
			retryable := d.errorRetryable(err)

			if !retryable {
				return err
			}

			d.logger.Info(fmt.Sprintf(
				"Retrying download due retryable error: %s",
				err.Error(),
			))

			continue
		}

		return nil
	}

	return err
}

// errorRetryable returns true if error indicates download can be retried.
// provided err must be non-nil
func (d Downloader) errorRetryable(err error) bool {
	if err == io.ErrUnexpectedEOF {
		d.logger.Info(fmt.Sprintf(
			"Received unexpected EOF error: %s",
			err.Error(),
		))
		return true
	}

	if netErr, ok := err.(net.Error); ok {
		if netErr.Temporary() {
			d.logger.Info(fmt.Sprintf(
				"Received temporary network error: %s",
				err.Error(),
			))
			return true
		}
	}

	return false
}
