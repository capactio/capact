//go:build integration
// +build integration

package e2e

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"testing"
	"time"

	enginegraphql "capact.io/capact/pkg/engine/api/graphql"
	engineclient "capact.io/capact/pkg/engine/client"
	"capact.io/capact/pkg/httputil"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/iosafety"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/yaml"
)

type GatewayConfig struct {
	Endpoint string
	Username string
	Password string
}

type Config struct {
	StatusEndpoints         []string
	IgnoredPodsNames        []string      `envconfig:"optional"`
	PollingInterval         time.Duration `envconfig:"default=2s"`
	PollingTimeout          time.Duration `envconfig:"default=5m"`
	Gateway                 GatewayConfig
	HubLocalDeployNamespace string `envconfig:"default=capact-system"`
	HubLocalDeployName      string `envconfig:"default=capact-hub-local"`
}

var (
	cfg                  Config
	originalGlobalPolicy enginegraphql.PolicyInput
)

var _ = BeforeSuite(func() {
	err := envconfig.Init(&cfg)
	Expect(err).ToNot(HaveOccurred())

	cli := getEngineGraphQLClient()
	originalPolicy, err := cli.GetPolicy(context.Background())

	rawPolicy, err := yaml.Marshal(originalPolicy)
	Expect(err).ToNot(HaveOccurred())

	input := enginegraphql.PolicyInput{}
	err = yaml.Unmarshal(rawPolicy, &input)
	Expect(err).ToNot(HaveOccurred())
	originalGlobalPolicy = input

	waitTillServiceEndpointsAreReady()
	waitTillDataIsPopulated()
})

var _ = AfterSuite(func() {
	cli := getEngineGraphQLClient()
	_, err := cli.UpdatePolicy(context.Background(), &originalGlobalPolicy)
	Expect(err).ToNot(HaveOccurred())
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
