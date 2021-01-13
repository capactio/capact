// +build integration

package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/util/podutils"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"
)

type Config struct {
	StatusEndpoints []string
}

func TestStatusEndpoints(t *testing.T) {
	var cfg Config
	err := envconfig.Init(&cfg)
	require.NoError(t, err)

	cli := httputil.NewClient(30*time.Second, true)

	for _, endpoint := range cfg.StatusEndpoints {
		resp, err := cli.Get(endpoint)
		assert.NoErrorf(t, err, "Get on %s", endpoint)

		err = iosafety.DrainReader(resp.Body)
		assert.NoError(t, err)
		err = resp.Body.Close()
		assert.NoError(t, err)
	}
}

const (
	// poll is how often to poll pods
	poll = 2 * time.Second
	// total number of pods that should be scheduled
	expectedNumberOfPods = 25
)

func TestAllPodsRunning(t *testing.T) {
	// create k8s client
	k8sCfg, err := config.GetConfig()
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(k8sCfg)

	numberOfPods := 0
	err = wait.PollImmediate(poll, time.Minute, func() (done bool, err error) {
		pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		numberOfPods = len(pods.Items)

		atLeastOneNotReady := false
		for _, p := range pods.Items {
			running, err := podRunningAndReady(t, &p)
			if err != nil {
				return false, err
			}

			if !running {
				atLeastOneNotReady = true
			}
		}
		return !atLeastOneNotReady, nil
	})

	require.NoError(t, err)
	assert.Equal(t, expectedNumberOfPods, numberOfPods, "got unexpected number of Pods in cluster")
}

func podRunningAndReady(t *testing.T, pod *v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodRunning:
		return podutils.IsPodReady(pod), nil
	}
	t.Logf("The status of Pod %s/%s is %s, waiting for it to be Running (with Ready = true)", pod.Namespace, pod.Name, pod.Status.Phase)
	return false, nil
}
