package pivnet_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix   = "/api/v2"
	productName = "some-product-name"
)

func TestPivnetResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PivnetResource Suite")
}
