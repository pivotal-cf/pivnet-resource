package sorter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSorter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sorter Suite")
}
