package extension

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

type ExtendedClient struct {
	pivnet.Client
	logger logger.Logger
}

func NewExtendedClient(config pivnet.ClientConfig, logger logger.Logger) ExtendedClient {
	return ExtendedClient{
		pivnet.NewClient(config, logger),
		logger,
	}
}

func (c ExtendedClient) ReleaseETag(productSlug string, releaseID int) (string, error) {
	url := fmt.Sprintf("/products/%s/releases/%d", productSlug, releaseID)

	resp, err := c.MakeRequest("GET", url, http.StatusOK, nil, nil)
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
