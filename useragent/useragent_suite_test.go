package useragent_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUserAgent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserAgent Suite")
}
