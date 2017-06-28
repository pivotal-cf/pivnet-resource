package validator

import (
	"fmt"

	"github.com/pivotal-cf/pivnet-resource/concourse"
)

type CheckValidator struct {
	input concourse.CheckRequest
}

func NewCheckValidator(input concourse.CheckRequest) *CheckValidator {
	return &CheckValidator{
		input: input,
	}
}

func (v CheckValidator) Validate() error {
	if v.input.Source.APIToken == "" && (v.input.Source.Username == "" || v.input.Source.Password == "") {
		return fmt.Errorf("%s must be provided", "username and password")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}
	return nil
}
