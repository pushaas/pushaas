package services_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

var logger *zap.Logger

func TestServices(t *testing.T) {
	logger = zaptest.NewLogger(t)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}
