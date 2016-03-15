package validator_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/validator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var vldtr validator.OutChecker

	BeforeEach(func() {
		vldtr = validator.OutChecker{}
	})

	DescribeTable("validating arguments",
		func(checker validator.OutChecker, input concourse.OutRequest, err error) {
			Expect(checker.Validate(input)).To(MatchError(err))
		},

		Entry("api_token missing", vldtr, concourse.OutRequest{}, errors.New("api_token must be provided")),
		// Entry("product_slug", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token"}}, errors.New("product_slug must be provided")),
		// Entry("version_file", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug"}}, errors.New("version_file must be provided")),
		// Entry("release_type_file", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug"}, Params: concourse.OutParams{VersionFile: "file"}}, errors.New("release_type_file must be provided")),
		// Entry("eula_slug_file", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug"}, Params: concourse.OutParams{VersionFile: "file", ReleaseTypeFile: "file"}}, errors.New("eula_slug_file must be provided")),
	)

	// Context("when file glob and file path prefix are both provided", func() {
	// 	DescribeTable("validating arguments",
	// 		func(checker validator.OutChecker, input concourse.OutRequest, err error) {
	// 			Expect(checker.Validate(input)).To(MatchError(err))
	// 		},

	// 		Entry("access_key_id missing", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug"}, Params: concourse.OutParams{VersionFile: "file", ReleaseTypeFile: "file", EulaSlugFile: "file", FileGlob: "glob glob", FilepathPrefix: "prefix"}}, errors.New("access_key_id must be provided")),
	// 		Entry("secret_access_key missing", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug", AccessKeyID: "key"}, Params: concourse.OutParams{VersionFile: "file", ReleaseTypeFile: "file", EulaSlugFile: "file", FileGlob: "glob glob", FilepathPrefix: "prefix"}}, errors.New("secret_access_key must be provided")),
	// 		Entry("file_glob missing", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug", AccessKeyID: "key", SecretAccessKey: "super-secret"}, Params: concourse.OutParams{VersionFile: "file", ReleaseTypeFile: "file", EulaSlugFile: "file", FilepathPrefix: "prefix"}}, errors.New("file_glob must be provided")),
	// 		Entry("s3_filepath_prefix missing", vldtr, concourse.OutRequest{Source: concourse.Source{APIToken: "token", ProductSlug: "slug", AccessKeyID: "key", SecretAccessKey: "super-secret"}, Params: concourse.OutParams{VersionFile: "file", ReleaseTypeFile: "file", EulaSlugFile: "file", FileGlob: "glob glob"}}, errors.New("s3_filepath_prefix must be provided")),
	// 	)
	// })
})
