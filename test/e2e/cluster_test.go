// +build integration

package e2e

import (
	"context"
	"fmt"
	"time"

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

	"projectvoltron.dev/voltron/pkg/httputil"

	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	client "projectvoltron.dev/voltron/pkg/engine/client"
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

var _ = Describe("Action E2E", func() {
	var engineClient *client.Client

	actionName := "e2e-test"
	ctx := context.Background()

	BeforeEach(func() {
		httpClient := httputil.NewClient(30*time.Second, true,
			httputil.WithBasicAuth(cfg.Gateway.Username, cfg.Gateway.Password))
		engineClient = client.New(cfg.Gateway.Endpoint, httpClient)
	})

	AfterEach(func() {
		// cleanup Action
		engineClient.DeleteAction(ctx, actionName)
	})

	Context("Action execution", func() {
		It("should have succeeded status after a passed workflow", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path: "cap.interface.voltron.e2e.passing-local",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusConditionReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusConditionSucceeded))
		})

		It("should have failed status after a failed workflow", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path: "cap.interface.voltron.e2e.failing-local",
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusConditionReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusConditionFailed))
		})
	})
})

func getActionStatusFunc(ctx context.Context, cl *client.Client, name string) func() (enginegraphql.ActionStatusCondition, error) {
	return func() (enginegraphql.ActionStatusCondition, error) {
		action, err := cl.GetAction(ctx, name)
		if err != nil {
			return "", err
		}
		return action.Status.Condition, err
	}
}

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
