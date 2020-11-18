package argo

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"go.uber.org/zap"

	"projectvoltron.dev/voltron/pkg/runner"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const testdataWorkflow = "./testdata/workflow.yaml"

func TestRunnerStartHappyPath(t *testing.T) {
	t.Run("Should create Argo Workflow if not exits", func(t *testing.T) {
		// given
		inputManifest, err := ioutil.ReadFile(testdataWorkflow)
		require.NoError(t, err)

		input := runner.StartInput{
			ExecCtx: runner.ExecutionContext{
				Name: "Rocket",
				Platform: runner.KubernetesPlatformConfig{
					Namespace: "argo-ns",
				},
			},
			Manifest: inputManifest,
		}

		var expWfSpec wfv1.WorkflowSpec
		require.NoError(t, yaml.Unmarshal(inputManifest, &expWfSpec))

		expOutStatus := Status{
			ArgoWorkflowRef: WorkflowRef{
				Name:      input.ExecCtx.Name,
				Namespace: input.ExecCtx.Platform.Namespace,
			},
		}

		fakeCli := fake.NewSimpleClientset()
		r := NewRunner(fakeCli.ArgoprojV1alpha1())

		// when
		out, err := r.Start(context.Background(), input)

		// then
		require.NoError(t, err)
		assert.Equal(t, expOutStatus, out.Status)

		gotWf, err := fakeCli.ArgoprojV1alpha1().Workflows(input.ExecCtx.Platform.Namespace).Get(input.ExecCtx.Name, metav1.GetOptions{})
		require.NoError(t, err)
		assert.EqualValues(t, expWfSpec, gotWf.Spec)
	})

	t.Run("Should skip Argo Workflow update if already exits", func(t *testing.T) {
		// given
		input := runner.StartInput{
			ExecCtx: runner.ExecutionContext{
				Name: "Rocket",
				Platform: runner.KubernetesPlatformConfig{
					Namespace: "argo-ns",
				},
			},
		}

		expOutStatus := Status{
			ArgoWorkflowRef: WorkflowRef{
				Name:      input.ExecCtx.Name,
				Namespace: input.ExecCtx.Platform.Namespace,
			},
		}

		wf := fixFinishedArgoWorkflow(t, input.ExecCtx.Name, input.ExecCtx.Platform.Namespace)
		fakeCli := fake.NewSimpleClientset(&wf)

		r := NewRunner(fakeCli.ArgoprojV1alpha1())
		r.InjectLogger(zap.NewNop())

		// when
		out, err := r.Start(context.Background(), input)

		// then
		require.NoError(t, err)
		assert.Equal(t, expOutStatus, out.Status)
	})
}

func TestRunnerStartFailure(t *testing.T) {
	t.Run("Should return error if dry run requested", func(t *testing.T) {
		// given
		input := runner.StartInput{
			ExecCtx: runner.ExecutionContext{
				DryRun: true,
			},
		}

		r := NewRunner(nil)

		// when
		out, err := r.Start(context.Background(), input)

		// then
		assert.EqualError(t, err, "DryRun support not implemented")
		assert.Nil(t, out)
	})

	t.Run("Should fail when manifest is malformed", func(t *testing.T) {
		// given
		input := runner.StartInput{
			Manifest: []byte("{malformed manifest"),
		}

		r := NewRunner(nil)

		// when
		out, err := r.Start(context.Background(), input)

		// then
		assert.EqualError(t, err, "while unmarshaling workflow spec: error converting YAML to JSON: yaml: line 1: did not find expected ',' or '}'")
		assert.Nil(t, out)
	})
}

func TestRunnerWaitForCompletion(t *testing.T) {
	// given
	input := runner.WaitForCompletionInput{
		ExecCtx: runner.ExecutionContext{
			Name: "Rocket",
			Platform: runner.KubernetesPlatformConfig{
				Namespace: "argo-ns",
			},
		},
	}

	wf := fixFinishedArgoWorkflow(t, input.ExecCtx.Name, input.ExecCtx.Platform.Namespace)

	fakeCli := fake.NewSimpleClientset(&wf)
	r := NewRunner(fakeCli.ArgoprojV1alpha1())

	// when
	err := waitForFunc(func(ctx context.Context) error {
		out, err := r.WaitForCompletion(ctx, input)
		assert.True(t, out.FinishedSuccessfully)
		return err
	}, 50*time.Millisecond)

	// then
	require.NoError(t, err)
}

func waitForFunc(fn func(ctx context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	wait, ctx := errgroup.WithContext(ctx)

	wait.Go(func() error {
		return fn(ctx)
	})

	return wait.Wait()
}

func fixFinishedArgoWorkflow(t *testing.T, name, ns string) wfv1.Workflow {
	t.Helper()

	rawWfSpec, err := ioutil.ReadFile(testdataWorkflow)
	require.NoError(t, err)

	var wfSpec wfv1.WorkflowSpec
	require.NoError(t, yaml.Unmarshal(rawWfSpec, &wfSpec))

	return wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: wfSpec,
		Status: wfv1.WorkflowStatus{
			FinishedAt: metav1.Now(),
			Phase:      wfv1.NodeSucceeded,
		},
	}
}
