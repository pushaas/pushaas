package apiV1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApiV1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApiV1 Suite")
}
