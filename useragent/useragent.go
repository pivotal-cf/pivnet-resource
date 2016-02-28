package useragent

import (
	"fmt"
	"os"
)

func UserAgent(version, containerType, productSlug string) string {
	atcExternalURL := os.Getenv("ATC_EXTERNAL_URL")
	buildPipelineName := os.Getenv("BUILD_PIPELINE_NAME")
	buildJobName := os.Getenv("BUILD_JOB_NAME")
	buildName := os.Getenv("BUILD_NAME")

	userAgent := fmt.Sprintf(
		"pivnet-resource/%s (%s/pipelines/%s/jobs/%s/builds/%s -- %s/%s)",
		version,
		atcExternalURL,
		buildPipelineName,
		buildJobName,
		buildName,
		productSlug,
		containerType,
	)
	return userAgent
}
