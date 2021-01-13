// +build integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/podutils"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"
	graphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client"
)

type GatewayConfig struct {
	Endpoint string
	Username string
	Password string
}

type Config struct {
	StatusEndpoints []string
	// total number of pods that should be scheduled
	ExpectedNumberOfRunningPods int
	IgnoredPodsNames            []string
	PollingInterval             time.Duration `envconfig:"default=2s"`
	PollingTimeout              time.Duration `envconfig:"default=1m"`
	Gateway                     GatewayConfig
}

func requireConfig(t *testing.T) Config {
	var cfg Config
	err := envconfig.Init(&cfg)
	require.NoError(t, err)
	return cfg
}

var _ = Describe("E2E", func() {
	cfg := Config{}
	ignoredPodsNames := map[string]struct{}{}

	BeforeSuite(func() {
		err := envconfig.Init(&cfg)
		Expect(err).ToNot(HaveOccurred())

		for _, n := range cfg.IgnoredPodsNames {
			ignoredPodsNames[n] = struct{}{}
		}
	})

	Describe("Voltron cluster health", func() {
		Context("Services status endpoint", func() {
			It("should be available", func() {
				cli := httputil.NewClient(30*time.Second, true)

				for _, endpoint := range cfg.StatusEndpoints {
					resp, err := cli.Get(endpoint)
					Expect(err).ToNot(HaveOccurred(), "Get on %s", endpoint)

					err = iosafety.DrainReader(resp.Body)
					Expect(err).ToNot(HaveOccurred())
					err = resp.Body.Close()
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})

		Context("Pods in cluster", func() {
			It("should be in running phase  (ignored kube-system)", func() {
				k8sCfg, err := config.GetConfig()
				Expect(err).ToNot(HaveOccurred())

				clientset, err := kubernetes.NewForConfig(k8sCfg)
				Expect(err).ToNot(HaveOccurred())
				Eventually(func() (int, error) {
					pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
						FieldSelector: fields.OneTermNotEqualSelector("metadata.namespace", "kube-system").String(),
					})
					if err != nil {
						return 0, err
					}

					atLeastOneNotReady := false
					numberOfRunningPods := 0
					for idx := range pods.Items {
						podName := strings.JoinQualifiedName(pods.Items[idx].Namespace, pods.Items[idx].Name)
						if _, skip := ignoredPodsNames[podName]; skip {
							continue
						}

						running := podRunningAndReadyOrFinished(&pods.Items[idx])
						if !running {
							atLeastOneNotReady = true
						} else {
							numberOfRunningPods++
						}
					}
					if atLeastOneNotReady {
						return 0, errors.New("detected not running pod(s)")
					}
					return numberOfRunningPods, nil
				}, cfg.PollingTimeout, cfg.PollingInterval).Should(Equal(cfg.ExpectedNumberOfRunningPods), "Got unexpected number of Pods in cluster")
			})
		})
	})
})

func podRunningAndReadyOrFinished(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return true
	case v1.PodRunning:
		ready := podutils.IsPodReady(pod)
		if !ready {
			log("The status of Pod %s/%s, waiting to be Ready", pod.Namespace, pod.Name)
		}
		return ready
	default:
		log("The status of Pod %s/%s is %s, waiting for it to be Running (with Ready = true)", pod.Namespace, pod.Name, pod.Status.Phase)
		return false
	}
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+format+"\n", args...)
}

func TestGatewayGetInterfaces(t *testing.T) {
	cfg := requireConfig(t)
	httpClient := httputil.NewClient(
		20*time.Second,
		true,
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	cli := client.NewClient(cfg.Gateway.Endpoint, httpClient)

	_, err := cli.GetInterfaces(context.Background())

	assert.NoError(t, err)
}

func TestOperationsOnTypeInstance(t *testing.T) {
	cfg := requireConfig(t)
	httpClient := httputil.NewClient(
		20*time.Second,
		true,
		httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password),
	)
	cli := client.NewClient(cfg.Gateway.Endpoint, httpClient)
	ctx := context.Background()

	createdTypeInstance, err := cli.CreateTypeInstance(ctx, &graphql.CreateTypeInstanceInput{
		TypeRef: &graphql.TypeReferenceInput{
			Path:     "com.voltron.ti",
			Revision: strPtr("0.1.0"),
		},
		Tags: []*graphql.TagReferenceInput{
			{
				Path:     "com.voltron.tag1",
				Revision: strPtr("0.1.0"),
			},
		},
		Value: map[string]interface{}{
			"foo": "bar",
		},
	})

	require.NoErrorf(t, err, "while creating TypeInstance: %v", err)
	require.Equal(t, map[string]interface{}{
		"foo": "bar",
	}, createdTypeInstance.Spec.Value)

	typeInstance, err := cli.GetTypeInstance(ctx, createdTypeInstance.Metadata.ID)
	require.NoErrorf(t, err, "while getting TypeInstance: %v", err)
	require.Equal(t, map[string]interface{}{
		"foo": "bar",
	}, typeInstance.Spec.Value)

	err = cli.DeleteTypeInstance(ctx, typeInstance.Metadata.ID)
	require.NoErrorf(t, err, "while deleting TypeInstance: %v", err)
}

func strPtr(s string) *string {
	return &s
}
