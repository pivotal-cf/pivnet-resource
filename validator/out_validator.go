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
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	return nil
}
