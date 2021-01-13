// +build integration

package e2e

import (
	"errors"
	"fmt"
	"time"

	"k8s.io/utils/strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/podutils"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"
)

const (
	// poll is how often to poll pods
	poll = 2 * time.Second
	// timeout for checking if all pods are running
	timeout = time.Minute
)

type Config struct {
	StatusEndpoints []string
	// total number of pods that should be scheduled
	ExpectedNumberOfRunningPods int `envconfig:"default=25"`
	IgnoredPodsNames            []string
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
			It("should be in running phase", func() {
				k8sCfg, err := config.GetConfig()
				Expect(err).ToNot(HaveOccurred())

				clientset, err := kubernetes.NewForConfig(k8sCfg)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() (int, error) {
					pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
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

						running := podRunningAndReady(&pods.Items[idx])
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
				}, timeout, poll).Should(Equal(cfg.ExpectedNumberOfRunningPods), "Got unexpected number of Pods in cluster")
			})
		})
	})
})

func podRunningAndReady(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodRunning:
		return podutils.IsPodReady(pod)
	}
	log("The status of Pod %s/%s is %s, waiting for it to be Running (with Ready = true)", pod.Namespace, pod.Name, pod.Status.Phase)
	return false
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+format+"\n", args...)
}
