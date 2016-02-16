package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Download(downloadDir string, downloadLinks map[string]string, token string) ([]string, error) {
	client := &http.Client{}

	fileNames := []string{}
	for fileName, downloadLink := range downloadLinks {
		req, err := http.NewRequest("POST", downloadLink, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if response.StatusCode == 451 {
			return nil, errors.New(fmt.Sprintf("the EULA has not been accepted for the file: %s", fileName))
		}

		if response.StatusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("pivnet returned an error code of %d for the file: %s", response.StatusCode, fileName))
		}

		downloadPath := filepath.Join(downloadDir, fileName)
		file, err := os.Create(downloadPath)
		if err != nil {
			return nil, err // not tested
		}

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return nil, err // not tested
		}

		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}
