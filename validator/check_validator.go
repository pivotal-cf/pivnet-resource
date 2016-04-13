package validator

import (
	"fmt"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
)

type checkValidator struct {
	input concourse.CheckRequest
}

func NewCheckValidator(input concourse.CheckRequest) Validator {
	return &checkValidator{
		input: input,
	}
}

func (v checkValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}
	return nil
}
