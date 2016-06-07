package validator

import (
	"fmt"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

type OutValidator struct {
	input concourse.OutRequest
}

func NewOutValidator(input concourse.OutRequest) OutValidator {
	return OutValidator{
		input: input,
	}
}

func (v OutValidator) Validate(skipFileCheck bool) error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if !skipFileCheck {
		if v.input.Params.VersionFile == "" {
			return fmt.Errorf("%s must be provided", "version_file")
		}

		if v.input.Params.ReleaseTypeFile == "" {
			return fmt.Errorf("%s must be provided", "release_type_file")
		}

		if v.input.Params.EULASlugFile == "" {
			return fmt.Errorf("%s must be provided", "eula_slug_file")
		}
	}

	if v.input.Params.FileGlob != "" || v.input.Params.FilepathPrefix != "" {
		if v.input.Source.AccessKeyID == "" {
			return fmt.Errorf("%s must be provided", "access_key_id")
		}

		if v.input.Source.SecretAccessKey == "" {
			return fmt.Errorf("%s must be provided", "secret_access_key")
		}

		if v.input.Params.FileGlob == "" {
			return fmt.Errorf("%s must be provided", "file glob")
		}

		if v.input.Params.FilepathPrefix == "" {
			return fmt.Errorf("%s must be provided", "s3_filepath_prefix")
		}
	}

	return nil
}
