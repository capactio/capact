//go:build integration
// +build integration

package e2e

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"testing"
	"time"

	engineclient "capact.io/capact/pkg/engine/client"
	"capact.io/capact/pkg/httputil"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/iosafety"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type GatewayConfig struct {
	Endpoint string
	Username string
	Password string
}

type GlobalPolicyConfig struct {
	Name      string `envconfig:"default=capact-engine-cluster-policy"`
	Namespace string `envconfig:"default=capact-system"`
}

type Config struct {
	StatusEndpoints         []string
	IgnoredPodsNames        []string      `envconfig:"optional"`
	PollingInterval         time.Duration `envconfig:"default=2s"`
	PollingTimeout          time.Duration `envconfig:"default=5m"`
	Gateway                 GatewayConfig
	ClusterPolicy           GlobalPolicyConfig
	HubLocalDeployNamespace string `envconfig:"default=capact-system"`
	HubLocalDeployName      string `envconfig:"default=capact-hub-local"`
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
	cli := httputil.NewClient(httputil.WithTLSInsecureSkipVerify(true))

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
	cli := getHubGraphQLClient()

	Eventually(func() (int, error) {
		ifaces, err := cli.ListInterfaces(context.Background())
		return len(ifaces), err
	}, cfg.PollingTimeout, cfg.PollingInterval).Should(BeNumerically(">", 1))
}

func getHubGraphQLClient() *hubclient.Client {
	httpClient := httputil.NewClient(
		httputil.WithTLSInsecureSkipVerify(true),
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	return hubclient.New(cfg.Gateway.Endpoint, httpClient)
}

func getEngineGraphQLClient() *engineclient.Client {
	httpClient := httputil.NewClient(
		httputil.WithTimeout(60*time.Second),
		httputil.WithTLSInsecureSkipVerify(true),
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	gqlClient := graphql.NewClient(cfg.Gateway.Endpoint, graphql.WithHTTPClient(httpClient))
	return engineclient.New(gqlClient)
}

func log(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+format+"\n", args...)
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}
