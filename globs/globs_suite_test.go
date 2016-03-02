package globs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGlobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Globs Suite")
}
