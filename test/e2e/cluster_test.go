// +build integration

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/podutils"
	"k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var _ = Describe("Cluster check", func() {
	ignoredPodsNames := map[string]struct{}{}

	BeforeEach(func() {
		for _, n := range cfg.IgnoredPodsNames {
			ignoredPodsNames[n] = struct{}{}
		}
	})

	Describe("Voltron cluster health", func() {
		Context("Services status endpoint", func() {
			It("should be available", func() {
				waitTillServiceEndpointsAreReady()
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
