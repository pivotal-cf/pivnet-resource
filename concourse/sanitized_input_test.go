package concourse_test

import (
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/pivotal-cf/go-pivnet"
  "github.com/pivotal-cf/pivnet-resource/concourse"
)

var _ = Describe("Sanitized Input", func() {
  Context("UaaEndpoint", func() {
    It("should return uaa specific endpoint for acceptance", func() {
      uaaEndpoint,err := concourse.UaaEndpoint("https://pivnet-acceptance.cfapps.io/")
      Expect(err).NotTo(HaveOccurred())
      Expect(uaaEndpoint).To(Equal("https://pivnet-acceptance-uaa.cfapps.io/"))
    })

    It("should return uaa specific endpoint for integration", func() {
      uaaEndpoint, err := concourse.UaaEndpoint("https://pivnet-integration.cfapps.io/")
      Expect(err).NotTo(HaveOccurred())
      Expect(uaaEndpoint).To(Equal("https://pivnet-integration-uaa.cfapps.io/"))
    })

    It("should return uaa specific endpoint for production", func() {
      uaaEndpoint, err := concourse.UaaEndpoint(pivnet.DefaultHost)
      Expect(err).NotTo(HaveOccurred())
      Expect(uaaEndpoint).To(Equal("https://pivnet-production-uaa.cfapps.io/"))
    })

    It("should handle invalid input", func() {
      _, err := concourse.UaaEndpoint("http://pivnet.io")
      Expect(err).To(HaveOccurred())
    })
  })
})