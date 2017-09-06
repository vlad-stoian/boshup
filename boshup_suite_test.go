package boshup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBoshup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boshup Suite")
}
