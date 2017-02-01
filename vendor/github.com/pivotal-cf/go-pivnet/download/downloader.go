package download

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"golang.org/x/sync/errgroup"
)

//go:generate counterfeiter -o ./fakes/ranger.go --fake-name Ranger . ranger
type ranger interface {
	BuildRange(contentLength int64) ([]Range, error)
}

//go:generate counterfeiter -o ./fakes/http_client.go --fake-name HTTPClient . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate counterfeiter -o ./fakes/bar.go --fake-name Bar . bar
type bar interface {
	SetTotal(contentLength int64)
	SetOutput(output io.Writer)
	Add(totalWritten int) int
	Kickoff()
	Finish()
	NewProxyReader(reader io.Reader) io.Reader
}

type Client struct {
	HTTPClient httpClient
	Ranger     ranger
	Bar        bar
}

func (c Client) Get(
	location *os.File,
	contentURL string,
	progressWriter io.Writer,
) error {
	req, err := http.NewRequest("HEAD", contentURL, nil)
	if err != nil {
		return fmt.Errorf("failed to construct HEAD request: %s", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make HEAD request: %s", err)
	}

	contentURL = resp.Request.URL.String()

	ranges, err := c.Ranger.BuildRange(resp.ContentLength)
	if err != nil {
		return fmt.Errorf("failed to construct range: %s", err)
	}

	c.Bar.SetOutput(progressWriter)
	c.Bar.SetTotal(resp.ContentLength)
	c.Bar.Kickoff()

	defer c.Bar.Finish()
	fileInfo, err := location.Stat()
	if err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	var g errgroup.Group
	for _, r := range ranges {
		byteRange := r

		fileWriter, err := os.OpenFile(location.Name(), os.O_RDWR, fileInfo.Mode())
		if err != nil {
			return fmt.Errorf("failed to write file: %s", err)
		}

		_, err = fileWriter.Seek(byteRange.Lower, 0)
		if err != nil {
			return fmt.Errorf("failed to write file: %s", err)
		}

		g.Go(func() error {
			err := c.retryableRequest(contentURL, byteRange.HTTPHeader, fileWriter)
			if err != nil {
				return fmt.Errorf("failed during retryable request: %s", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (c Client) retryableRequest(contentURL string, rangeHeader http.Header, fileWriter *os.File) (error) {
	currentURL := contentURL
	defer fileWriter.Close()
Retry:
	req, err := http.NewRequest("GET", currentURL, nil)
	if err != nil {
		return err
	}

	req.Header = rangeHeader

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Temporary() {
				goto Retry
			}
		}

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("during GET unexpected status code was returned: %d", resp.StatusCode)
	}

	var proxyReader io.Reader
	proxyReader = c.Bar.NewProxyReader(resp.Body)

	_, err = io.Copy(fileWriter, proxyReader)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			goto Retry
		}
		return fmt.Errorf("failed to write file: %s", err)
	}

	return nil
}
