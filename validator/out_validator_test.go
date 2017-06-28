package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/validator"
)

var _ = Describe("Out Validator", func() {
	var (
		accessKeyID     string
		secretAccessKey string

		apiToken         string
		productSlug      string
		fileGlob         string
		s3FilepathPrefix string

		username string
		password string

		outRequest concourse.OutRequest
		v          *validator.OutValidator
	)

	BeforeEach(func() {
		accessKeyID = "some-access-key"
		secretAccessKey = "some-secret-access-key"
		apiToken = "some-api-token"
		productSlug = "some-product"
		username = "username"
		password = "password"

		fileGlob = ""
		s3FilepathPrefix = ""
	})

	JustBeforeEach(func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Username:        username,
				Password:        password,
				APIToken:        apiToken,
				ProductSlug:     productSlug,
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
			Params: concourse.OutParams{
				FileGlob:       fileGlob,
				FilepathPrefix: s3FilepathPrefix,
			},
		}

		v = validator.NewOutValidator(outRequest)
	})

	It("returns without error", func() {
		Expect(v.Validate()).NotTo(HaveOccurred())
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

	Context("when file glob is not provided", func() {
		BeforeEach(func() {
			fileGlob = ""
		})

		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when s3 filepath prefix is not provided", func() {
		BeforeEach(func() {
			s3FilepathPrefix = ""
		})

		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when file glob is present", func() {
		BeforeEach(func() {
			fileGlob = "some-file-glob"
		})

		Context("when s3 filepath prefix is not provided", func() {
			BeforeEach(func() {
				s3FilepathPrefix = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*s3_filepath_prefix.*provided"))
			})
		})

		Context("when no aws access key id is provided", func() {
			BeforeEach(func() {
				accessKeyID = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*access_key_id.*provided"))
			})
		})

		Context("when no aws secret access key is provided", func() {
			BeforeEach(func() {
				secretAccessKey = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*secret_access_key.*provided"))
			})
		})
	})

	Context("when filepath prefix is present", func() {
		BeforeEach(func() {
			s3FilepathPrefix = "some-filepath-prefix"
		})

		Context("when file glob is not provided", func() {
			BeforeEach(func() {
				fileGlob = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*file glob.*provided"))
			})
		})

		Context("when no aws access key id is provided", func() {
			BeforeEach(func() {
				accessKeyID = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*access_key_id.*provided"))
			})
		})

		Context("when no aws secret access key is provided", func() {
			BeforeEach(func() {
				secretAccessKey = ""
			})

			It("returns an error", func() {
				err := v.Validate()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*secret_access_key.*provided"))
			})
		})
	})
})
