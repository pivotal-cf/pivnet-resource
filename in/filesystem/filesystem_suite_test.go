package filesystem_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	productSlug = "some-product-name"
)

func TestInFilesystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "In Filesystem Suite")
}
