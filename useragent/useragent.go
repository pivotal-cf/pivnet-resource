package useragent

import (
	"fmt"
	"os"
)

func UserAgent(version, containerType, productSlug string) string {
	atcExternalURL := os.Getenv("ATC_EXTERNAL_URL")

	// check containers
	resourceName := os.Getenv("RESOURCE_NAME")
	pipelineName := os.Getenv("PIPELINE_NAME")

	// check container
	if resourceName != "" {
		return fmt.Sprintf(
			"pivnet-resource/%s (%s/pipelines/%s/resources/%s -- %s)",
			version,
			atcExternalURL,
			pipelineName,
			resourceName,
			containerType,
		)
	}

	// in/out containers
	buildPipelineName := os.Getenv("BUILD_PIPELINE_NAME")
	buildJobName := os.Getenv("BUILD_JOB_NAME")
	buildName := os.Getenv("BUILD_NAME")

	return fmt.Sprintf(
		"pivnet-resource/%s (%s/pipelines/%s/jobs/%s/builds/%s -- %s/%s)",
		version,
		atcExternalURL,
		buildPipelineName,
		buildJobName,
		buildName,
		productSlug,
		containerType,
	)
}
