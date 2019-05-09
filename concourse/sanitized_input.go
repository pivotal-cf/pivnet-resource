package concourse

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/go-pivnet"
)

func SanitizedSource(source Source) map[string]string {
	s := make(map[string]string)

	if source.APIToken != "" {
		s[source.APIToken] = "***REDACTED-PIVNET_API_TOKEN***"
	}

	return s
}

func UaaEndpoint(endpoint string) (string, error) {
	if endpoint == pivnet.DefaultHost {
		return "https://pivnet-production-uaa.cfapps.io/", nil
	}

	endpointDelimiter := ".cfapps.io"
	if !strings.Contains(endpoint, endpointDelimiter){
		return "", fmt.Errorf("%s is not a valid endpoint", endpoint)
	}

	uaaEndpoint := strings.Replace(endpoint, endpointDelimiter, fmt.Sprintf("-uaa%s", endpointDelimiter), 1)

	return uaaEndpoint, nil
}