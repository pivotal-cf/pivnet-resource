package uaa_test

import (
. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"

"testing"
)

func TestUAA(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UAA Suite")
}
