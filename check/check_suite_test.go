package check_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix   = "/api/v2"
	productSlug = "some-product-name"
)

func TestCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Check Suite")
}
