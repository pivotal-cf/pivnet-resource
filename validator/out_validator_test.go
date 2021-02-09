package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"
	"github.com/pivotal-cf/pivnet-resource/v3/validator"
)

var _ = Describe("Out Validator", func() {
	var (

		apiToken         string
		productSlug      string
		fileGlob         string

		outRequest concourse.OutRequest
		v          *validator.OutValidator
	)

	BeforeEach(func() {
		apiToken = "some-api-token"
		productSlug = "some-product"

		fileGlob = ""
	})

	JustBeforeEach(func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        apiToken,
				ProductSlug:     productSlug,
			},
			Params: concourse.OutParams{
				FileGlob:       fileGlob,
			},
		}

		v = validator.NewOutValidator(outRequest)
	})

	It("returns without error", func() {
		Expect(v.Validate()).NotTo(HaveOccurred())
	})

	Context("when neither UAA refresh token nor legacy API token are provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("api_token must be provided"))
		})
	})

	Context("when UAA reresh token or legacy API token is provided", func() {
		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
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

	Context("when file glob is not provided", func() {
		BeforeEach(func() {
			fileGlob = ""
		})

		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
