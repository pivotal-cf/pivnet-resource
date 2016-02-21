package check

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/versions"
)

type CheckCommand struct {
	logger  logger.Logger
	logFile *os.File
	version string
}

func NewCheckCommand(
	version string,
	logger logger.Logger,
	logFile *os.File,
) *CheckCommand {
	return &CheckCommand{
		logger:  logger,
		logFile: logFile,
		version: version,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	logDir := filepath.Dir(c.logFile.Name())
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "pivnet-resource-check.log*"))
	if err != nil {
		c.logger.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFile.Name()) {
			c.logger.Debugf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				c.logger.Debugf("Exiting with error: %v\n", err)
				log.Fatalln(err)
			}
		}
	}

	if input.Source.APIToken == "" {
		return nil, fmt.Errorf("%s must be provided", "api_token")
	}

	if input.Source.ProductSlug == "" {
		return nil, fmt.Errorf("%s must be provided", "product_slug")
	}

	c.logger.Debugf("Received input: %+v\n", input)

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.Endpoint
	}

	clientConfig := pivnet.NewClientConfig{
		Endpoint:  endpoint,
		Token:     input.Source.APIToken,
		UserAgent: fmt.Sprintf("pivnet-resource/%s", c.version),
	}
	client := pivnet.NewClient(
		clientConfig,
		c.logger,
	)

	c.logger.Debugf("Getting all product versions\n")

	allVersions, err := client.ProductVersions(input.Source.ProductSlug)
	if err != nil {
		c.logger.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	c.logger.Debugf("All known versions: %+v\n", allVersions)

	if len(allVersions) == 0 {
		return concourse.CheckResponse{}, nil
	}

	newVersions, err := versions.Since(allVersions, input.Version.ProductVersion)
	if err != nil {
		c.logger.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	c.logger.Debugf("New versions: %+v\n", newVersions)

	reversedVersions, err := versions.Reverse(newVersions)
	if err != nil {
		c.logger.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	var out concourse.CheckResponse
	for _, v := range reversedVersions {
		out = append(out, concourse.Version{ProductVersion: v})
	}

	if len(out) == 0 {
		out = append(out, concourse.Version{ProductVersion: allVersions[0]})
	}

	c.logger.Debugf("Returning output: %+v\n", out)

	return out, nil
}

func (c *CheckCommand) mustBeNonEmpty(input string, key string) {
	if input == "" {
		c.logger.Debugf("Exiting with error: %s must be provided\n", key)
		log.Fatalf("%s must be provided\n", key)
	}
}
