package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"
)

var _ = Describe("Out Validator", func() {
	var (
		accessKeyID     string
		secretAccessKey string

		versionFile      string
		apiToken         string
		productSlug      string
		fileGlob         string
		s3FilepathPrefix string
		releaseTypeFile  string
		eulaSlugFile     string

		outRequest concourse.OutRequest
		v          validator.Validator
	)

	BeforeEach(func() {
		accessKeyID = "some-access-key"
		secretAccessKey = "some-secret-access-key"
		apiToken = "some-api-token"
		productSlug = "some-product"

		versionFile = "some-version-file"
		releaseTypeFile = "some-release-type-file"
		eulaSlugFile = "some-eula-slug-file"

		fileGlob = ""
		s3FilepathPrefix = ""
	})

	JustBeforeEach(func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				APIToken:        apiToken,
				ProductSlug:     productSlug,
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
			Params: concourse.OutParams{
				VersionFile:     versionFile,
				FileGlob:        fileGlob,
				FilepathPrefix:  s3FilepathPrefix,
				ReleaseTypeFile: releaseTypeFile,
				EulaSlugFile:    eulaSlugFile,
			},
		}

		v = validator.NewOutValidator(outRequest, false)
	})

	Context("when skipping file checks", func() {
		BeforeEach(func() {
			versionFile = ""
			eulaSlugFile = ""
			releaseTypeFile = ""
		})

		It("ignores the fact that files are not provided", func() {
			v = validator.NewOutValidator(outRequest, true)
			Expect(v.Validate()).NotTo(HaveOccurred())
		})
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

	Context("when version file is not provided", func() {
		BeforeEach(func() {
			versionFile = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*version_file.*provided"))
		})
	})

	Context("when release_type file is not provided", func() {
		BeforeEach(func() {
			releaseTypeFile = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*release_type_file.*provided"))
		})
	})

	Context("when eula_slug file is not provided", func() {
		BeforeEach(func() {
			eulaSlugFile = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*eula_slug_file.*provided"))
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
