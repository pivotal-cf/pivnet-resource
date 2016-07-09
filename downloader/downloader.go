package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//go:generate counterfeiter . Downloader

type Downloader interface {
	Download(downloadLinks map[string]string) ([]string, error)
}

type downloader struct {
	apiToken    string
	downloadDir string
	logger      *log.Logger
}

func NewDownloader(apiToken string, downloadDir string, logger *log.Logger) Downloader {
	return &downloader{
		apiToken:    apiToken,
		downloadDir: downloadDir,
		logger:      logger,
	}
}

func (d downloader) Download(downloadLinks map[string]string) ([]string, error) {
	d.logger.Println("Ensuring download directory exists")

	err := os.MkdirAll(d.downloadDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for fileName, downloadLink := range downloadLinks {
		downloadPath, err := d.download(fileName, downloadLink)
		if err != nil {
			return nil, err

		}
		fileNames = append(fileNames, downloadPath)

	}

	return fileNames, nil
}

func (d downloader) download(fileName string, downloadLink string) (string, error) {
	downloadPath := filepath.Join(d.downloadDir, fileName)

	req, err := http.NewRequest("POST", downloadLink, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", d.apiToken))

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode == http.StatusUnavailableForLegalReasons {
		return "", fmt.Errorf("the EULA has not been accepted for the file: %s", fileName)
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pivnet returned an error code of %d for the file: %s", response.StatusCode, fileName)
	}

	file, err := os.Create(downloadPath)
	if err != nil {
		return "", err // not tested
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err // not tested
	}

	return downloadPath, nil
}
