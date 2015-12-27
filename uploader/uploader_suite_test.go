package uploader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUploader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uploader Suite")
}
