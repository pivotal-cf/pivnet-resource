package pivnet_resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPivnetResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PivnetResource Suite")
}
