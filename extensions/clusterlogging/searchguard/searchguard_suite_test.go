package searchguard

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSearchGuard(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SearchGuard Suite")
}
