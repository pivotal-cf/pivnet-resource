package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
)

var _ = Describe("In Validator", func() {
	var (
		inRequest concourse.InRequest
		v         validator.Validator

		apiToken       string
		productSlug    string
		productVersion string
	)

	BeforeEach(func() {
		apiToken = "some-api-token"
		productSlug = "some-productSlug"
		productVersion = "some-product-version"
	})

	JustBeforeEach(func() {
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:    apiToken,
				ProductSlug: productSlug,
			},
			Params: concourse.InParams{},
			Version: concourse.Version{
				ProductVersion: productVersion,
			},
		}

		v = validator.NewInValidator(inRequest)
	})

	It("returns without error", func() {
		// Expect(v.Validate()).NotTo(HaveOccurred())
		Expect(v.Validate()).To(BeNil())
	})

	Context("when no api token is provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			// Expect(err).To(HaveOccurred())
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
		})
	})

	Context("when no product slug is provided", func() {
		BeforeEach(func() {
			productSlug = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			// Expect(err).To(HaveOccurred())
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
		})
	})

	Context("when no product version is provided", func() {
		BeforeEach(func() {
			productVersion = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*product_version.*provided"))
		})
	})
})
