//go:build integration
// +build integration

package e2e

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/podutils"
	k8sstrings "k8s.io/utils/strings"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var _ = Describe("Cluster check", func() {
	ignoredPodsNames := map[string]struct{}{}

	BeforeEach(func() {
		for _, n := range cfg.IgnoredPodsNames {
			ignoredPodsNames[n] = struct{}{}
		}
	})

	Describe("Capact cluster health", func() {
		Context("Pods in cluster", func() {
			It("should be in running phase (ignored Namespace: [kube-system, default], ignored evicted Pods)", func() {
				ctx := context.Background()
				k8sCfg, err := config.GetConfig()
				Expect(err).ToNot(HaveOccurred())

				clientset, err := kubernetes.NewForConfig(k8sCfg)
				Expect(err).ToNot(HaveOccurred())
				Eventually(func() error {
					pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{
						FieldSelector: fields.AndSelectors(
							fields.OneTermNotEqualSelector("metadata.namespace", "kube-system"),
							fields.OneTermNotEqualSelector("metadata.namespace", "default"),
						).String(),
					})
					if err != nil {
						return err
					}

					var notReadyPods []string
					for idx := range pods.Items {
						podName := k8sstrings.JoinQualifiedName(pods.Items[idx].Namespace, pods.Items[idx].Name)
						if _, skip := ignoredPodsNames[podName]; skip {
							continue
						}

						running := podRunningAndReadyOrFinished(&pods.Items[idx])
						if !running {
							notReadyPods = append(notReadyPods, podName)
						}
					}
					if len(notReadyPods) > 0 {
						return errors.Errorf("detected not running pod(s): %s", strings.Join(notReadyPods, ", "))
					}
					return nil
				}, cfg.PollingTimeout, cfg.PollingInterval).ShouldNot(HaveOccurred())
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
		// Ignore evicted pods which are in Failed state because of graceful node shutdown.
		// see: https://kubernetes.io/docs/concepts/architecture/nodes/#graceful-node-shutdown
		if strings.EqualFold(pod.Status.Reason, "Shutdown") {
			return true
		}
		log("The status of Pod %s/%s is %s, waiting for it to be Running (with Ready = true)", pod.Namespace, pod.Name, pod.Status.Phase)
		return false
	}
}
