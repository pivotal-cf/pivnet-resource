package useragent_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/useragent"
)

var _ = Describe("UserAgent", func() {
	It("creates user agent string from environment variables", func() {
		version := "0.2.1"
		containerType := "get"
		productSlug := "my-product"

		atcExternalURL := "https://some-external-url"
		buildPipelineName := "some-pipeline"
		buildJobName := "build-job-name"
		buildName := "build-name"

		err := os.Setenv("ATC_EXTERNAL_URL", atcExternalURL)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("BUILD_PIPELINE_NAME", buildPipelineName)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("BUILD_JOB_NAME", buildJobName)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("BUILD_NAME", buildName)
		Expect(err).NotTo(HaveOccurred())

		userAgentString := useragent.UserAgent(version, containerType, productSlug)
		Expect(userAgentString).To(Equal(
			"pivnet-resource/0.2.1 (https://some-external-url/pipelines/some-pipeline/jobs/build-job-name/builds/build-name -- my-product/get)",
		))
	})
})
