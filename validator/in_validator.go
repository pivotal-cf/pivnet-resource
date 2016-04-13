package validator

import (
	"fmt"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

type inValidator struct {
	input concourse.InRequest
}

func NewInValidator(input concourse.InRequest) Validator {
	return &inValidator{
		input: input,
	}
}

func (v inValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if v.input.Version.ProductVersion == "" {
		return fmt.Errorf("%s must be provided", "product_version")
	}

	return nil
}
