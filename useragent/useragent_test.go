package useragent_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/useragent"
)

var _ = Describe("UserAgent", func() {
	var (
		version     string
		productSlug string

		containerType string
	)

	BeforeEach(func() {
		version = "0.2.1"
		productSlug = "my-product"
	})

	Context("when check container environment variables are present", func() {
		var (
			externalURL = "https://some-external-url"

			resourceName string
			pipelineName string
		)

		BeforeEach(func() {
			containerType = "check"

			externalURL = "https://some-external-url"

			resourceName = "some-resource"
			pipelineName = "some-pipeline"

			err := os.Setenv("EXTERNAL_URL", externalURL)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("PIPELINE_NAME", pipelineName)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("RESOURCE_NAME", resourceName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Unsetenv("PIPELINE_NAME")
			Expect(err).NotTo(HaveOccurred())

			err = os.Unsetenv("RESOURCE_NAME")
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates user agent string from environment variables", func() {
			userAgentString := useragent.UserAgent(version, containerType, productSlug)

			Expect(userAgentString).To(Equal(
				"pivnet-resource/0.2.1 (https://some-external-url/pipelines/some-pipeline/resources/some-resource -- some-resource/check)",
			))
		})

		Context("when resource name has an emoji", func() {
			JustBeforeEach(func() {
				err := os.Setenv("RESOURCE_NAME", "some-resource-☢")
				Expect(err).NotTo(HaveOccurred())
			})

			JustAfterEach(func() {
				err := os.Unsetenv("RESOURCE_NAME")
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates user agent string from environment variables with go escaped sequence for emojis", func() {
				userAgentString := useragent.UserAgent(version, containerType, productSlug)

				Expect(userAgentString).To(Equal(
					"pivnet-resource/0.2.1 (https://some-external-url/pipelines/some-pipeline/resources/some-resource-\\u2622 -- some-resource-\\u2622/check)",
				))
			})
		})
	})

	Context("when in/out container environment variables are present", func() {
		var (
			atcExternalURL string

			buildPipelineName string
			buildJobName      string
			buildName         string
		)

		BeforeEach(func() {
			containerType = "get"

			atcExternalURL = "https://some-external-url"

			buildPipelineName = "some-pipeline"
			buildJobName = "build-job-name"
			buildName = "build-name"

			err := os.Setenv("ATC_EXTERNAL_URL", atcExternalURL)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("BUILD_PIPELINE_NAME", buildPipelineName)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("BUILD_JOB_NAME", buildJobName)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("BUILD_NAME", buildName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Unsetenv("BUILD_PIPELINE_NAME")
			Expect(err).NotTo(HaveOccurred())

			err = os.Unsetenv("BUILD_JOB_NAME")
			Expect(err).NotTo(HaveOccurred())

			err = os.Unsetenv("BUILD_NAME")
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates user agent string from environment variables", func() {
			userAgentString := useragent.UserAgent(version, containerType, productSlug)

			Expect(userAgentString).To(Equal(
				"pivnet-resource/0.2.1 (https://some-external-url/pipelines/some-pipeline/jobs/build-job-name/builds/build-name -- my-product/get)",
			))
		})

		Context("when job name has an emoji", func() {
			JustBeforeEach(func() {
				err := os.Setenv("BUILD_JOB_NAME", "job-☢")
				Expect(err).NotTo(HaveOccurred())
			})

			JustAfterEach(func() {
				err := os.Unsetenv("BUILD_JOB_NAME")
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates user agent string from environment variables with go escaped sequence for emojis", func() {
				userAgentString := useragent.UserAgent(version, containerType, productSlug)

				Expect(userAgentString).To(Equal(
					"pivnet-resource/0.2.1 (https://some-external-url/pipelines/some-pipeline/jobs/job-\\u2622/builds/build-name -- my-product/get)",
				))
			})
		})
	})
})
