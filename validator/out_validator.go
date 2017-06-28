package validator

import (
	"fmt"

	"github.com/pivotal-cf/pivnet-resource/concourse"
)

type OutValidator struct {
	input concourse.OutRequest
}

func NewOutValidator(input concourse.OutRequest) *OutValidator {
	return &OutValidator{
		input: input,
	}
}

func (v OutValidator) Validate() error {
	if v.input.Source.APIToken == "" && (v.input.Source.Username == "" || v.input.Source.Password == "") {
		return fmt.Errorf("%s must be provided", "username and password")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
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
