package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
)

var _ = Describe("Check Validator", func() {
	var (
		checkRequest concourse.CheckRequest
		v            *validator.CheckValidator

		apiToken    string
		productSlug string
	)

	BeforeEach(func() {
		apiToken = "some-api-token"
		productSlug = "some-productSlug"
	})

	JustBeforeEach(func() {
		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    apiToken,
				ProductSlug: productSlug,
			},
		}
		v = validator.NewCheckValidator(checkRequest)
	})

	It("returns without error", func() {
		err := v.Validate()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no api token is provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(".*api_token.*provided"))
		})
	})

	Context("when no product slug is provided", func() {
		BeforeEach(func() {
			productSlug = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
		})
	})
})
