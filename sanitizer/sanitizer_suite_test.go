package sanitizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSanitizer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanitizer Suite")
}
