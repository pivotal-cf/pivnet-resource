package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/validator"
)

var _ = Describe("In Validator", func() {
	var (
		inRequest concourse.InRequest
		v         *validator.InValidator

		username    string
		password    string
		apiToken    string
		productSlug string
		version     string
	)

	BeforeEach(func() {
		username = "username"
		password = "password"
		apiToken = "some-api-token"
		productSlug = "some-productSlug"
		version = "some-product-version"
	})

	JustBeforeEach(func() {
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Username:    username,
				Password:    password,
				APIToken:    apiToken,
				ProductSlug: productSlug,
			},
			Params: concourse.InParams{},
			Version: concourse.Version{
				ProductVersion: version,
			},
		}

		v = validator.NewInValidator(inRequest)
	})

	It("returns without error", func() {
		err := v.Validate()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no api token is provided but uaa credentials are provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when uaa credentials and api token are not provided", func() {
		BeforeEach(func() {
			username = ""
			password = ""
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("username and password must be provided"))
		})
	})

	Context("when username is provided but password is not provided", func() {
		BeforeEach(func() {
			username = ""
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("username and password must be provided"))
		})
	})

	Context("when uaa credentials are not provided but api token is provided", func() {
		BeforeEach(func() {
			username = ""
			password = ""
		})

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

	Context("when no product version is provided", func() {
		BeforeEach(func() {
			version = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*product_version.*provided"))
		})
	})
})
