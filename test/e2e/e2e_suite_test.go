// +build integration

package e2e

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"
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
	PollingTimeout              time.Duration `envconfig:"default=2m"`
	Gateway                     GatewayConfig
}

var cfg Config

var _ = BeforeSuite(func() {
	err := envconfig.Init(&cfg)
	Expect(err).ToNot(HaveOccurred())

	waitTillServiceEndpointsAreReady()
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

func waitTillServiceEndpointsAreReady() {
	cli := httputil.NewClient(30*time.Second, true)

	for _, endpoint := range cfg.StatusEndpoints {
		Eventually(func() error {
			resp, err := cli.Get(endpoint)
			if err != nil {
				return errors.Wrapf(err, "while GET on %s", endpoint)
			}

			err = iosafety.DrainReader(resp.Body)
			if err != nil {
				return nil
			}

			err = resp.Body.Close()
			return err
		}, cfg.PollingTimeout, cfg.PollingInterval).ShouldNot(HaveOccurred())
	}
}
