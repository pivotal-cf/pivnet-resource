package validator

import (
	"fmt"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

type OutChecker struct{}

func (o OutChecker) Validate(input concourse.OutRequest) error {
	if input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}
	if input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}
	if input.Params.VersionFile == "" {
		return fmt.Errorf("%s must be provided", "version_file")
	}
	if input.Params.ReleaseTypeFile == "" {
		return fmt.Errorf("%s must be provided", "release_type_file")
	}
	if input.Params.EulaSlugFile == "" {
		return fmt.Errorf("%s must be provided", "eula_slug_file")
	}

	skipUpload := input.Params.FileGlob == "" && input.Params.FilepathPrefix == ""

	if !skipUpload {
		if input.Source.AccessKeyID == "" {
			return fmt.Errorf("%s must be provided", "access_key_id")
		}

		if input.Source.SecretAccessKey == "" {
			return fmt.Errorf("%s must be provided", "secret_access_key")
		}

		if input.Params.FileGlob == "" {
			return fmt.Errorf("%s must be provided", "file_glob")
		}

		if input.Params.FilepathPrefix == "" {
			return fmt.Errorf("%s must be provided", "s3_filepath_prefix")
		}
	}
	return nil
}
