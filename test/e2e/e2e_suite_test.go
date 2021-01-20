// +build integration

package e2e

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
)

type GatewayConfig struct {
	Endpoint string
	Username string
	Password string
}

type Config struct {
	StatusEndpoints []string
	// total number of pods that should be scheduled
	ExpectedNumberOfRunningPods int `envconfig:"default=25"`
	IgnoredPodsNames            []string
	PollingInterval             time.Duration `envconfig:"default=2s"`
	PollingTimeout              time.Duration `envconfig:"default=1m"`
	Gateway                     GatewayConfig
}

var cfg Config

var _ = BeforeSuite(func() {
	err := envconfig.Init(&cfg)
	Expect(err).ToNot(HaveOccurred())
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}
