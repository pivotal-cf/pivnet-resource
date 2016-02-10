package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		downloadPath := filepath.Join(downloadDir, fileName)
		err = ioutil.WriteFile(downloadPath, responseBody, 0666)
		if err != nil {
			return nil, err
		}

		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}
