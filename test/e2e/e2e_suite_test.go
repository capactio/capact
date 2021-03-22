// +build integration

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	engineclient "projectvoltron.dev/voltron/pkg/engine/client"
	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
)

type GatewayConfig struct {
	Endpoint string
	Username string
	Password string
}

type ClusterPolicyConfig struct {
	Name      string `envconfig:"default=voltron-engine-cluster-policy"`
	Namespace string `envconfig:"default=voltron-system"`
}

type Config struct {
	StatusEndpoints []string
	// total number of pods that should be scheduled
	ExpectedNumberOfRunningPods int           `envconfig:"default=25"`
	IgnoredPodsNames            []string      `envconfig:"optional"`
	PollingInterval             time.Duration `envconfig:"default=2s"`
	PollingTimeout              time.Duration `envconfig:"default=5m"`
	Gateway                     GatewayConfig
	ClusterPolicy               ClusterPolicyConfig
	OCHLocalDeployNamespace     string `envconfig:"default=voltron-system"`
	OCHLocalDeployName          string `envconfig:"default=voltron-och-local"`
}

var cfg Config

var _ = BeforeSuite(func() {
	err := envconfig.Init(&cfg)
	Expect(err).ToNot(HaveOccurred())

	waitTillServiceEndpointsAreReady()
	waitTillDataIsPopulated()
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
		}, 5*cfg.PollingTimeout, cfg.PollingInterval).ShouldNot(HaveOccurred())
	}
}

func waitTillDataIsPopulated() {
	cli := getOCHGraphQLClient()

	Eventually(func() (int, error) {
		ifaces, err := cli.ListInterfacesMetadata(context.Background())
		return len(ifaces), err
	}, cfg.PollingTimeout, cfg.PollingInterval).Should(BeNumerically(">", 1))
}

func getOCHGraphQLClient() *ochclient.Client {
	httpClient := httputil.NewClient(
		30*time.Second,
		true,
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	return ochclient.NewClient(cfg.Gateway.Endpoint, httpClient)
}

func getEngineGraphQLClient() *engineclient.Client {
	httpClient := httputil.NewClient(
		30*time.Second,
		true,
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	return engineclient.New(cfg.Gateway.Endpoint, httpClient)
}

func log(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+format+"\n", args...)
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}
