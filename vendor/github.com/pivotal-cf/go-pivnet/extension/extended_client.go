package extension

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/pivotal-cf/go-pivnet/logger"
)

//go:generate counterfeiter . Client
type Client interface {
	MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error)
	CreateRequest(method string, url string, body io.Reader) (*http.Request, error)
}

type ExtendedClient struct {
	c      Client
	logger logger.Logger
}

func NewExtendedClient(client Client, logger logger.Logger) ExtendedClient {
	return ExtendedClient{
		c:      client,
		logger: logger,
	}
}

func (c ExtendedClient) ReleaseETag(productSlug string, releaseID int) (string, error) {
	url := fmt.Sprintf("/products/%s/releases/%d", productSlug, releaseID)

	resp, err := c.c.MakeRequest("GET", url, http.StatusOK, nil, nil)
	if err != nil {
		return "", err
	}

	rawEtag := resp.Header.Get("ETag")

	if rawEtag == "" {
		return "", fmt.Errorf("ETag header not present")
	}

	c.logger.Debug("Received ETag", logger.Data{"etag": rawEtag})

	// Weak ETag looks like: W/"my-etag"; strong ETag looks like: "my-etag"
	splitRawEtag := strings.SplitN(rawEtag, `"`, -1)

	if len(splitRawEtag) < 2 {
		return "", fmt.Errorf("ETag header malformed: %s", rawEtag)
	}

	etag := splitRawEtag[1]
	return etag, nil
}

func (c ExtendedClient) DownloadFile(writer io.Writer, downloadLink string) error {
	c.logger.Debug("Downloading file", logger.Data{"downloadLink": downloadLink})

	req, err := c.c.CreateRequest(
		"POST",
		downloadLink,
		nil,
	)
	if err != nil {
		return err
	}

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}

	c.logger.Debug("Making request", logger.Data{"request": string(reqBytes)})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnavailableForLegalReasons {
		return errors.New(fmt.Sprintf("the EULA has not been accepted for the file: %s", downloadLink))
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("pivnet returned an error code of %d for the file: %s", resp.StatusCode, downloadLink))
	}

	c.logger.Debug("Copying body", logger.Data{"downloadLink": downloadLink})

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
