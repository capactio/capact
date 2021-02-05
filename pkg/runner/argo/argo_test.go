package argo

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"projectvoltron.dev/voltron/pkg/runner"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/fake"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/scheme"
	argoprojv1alpha1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	fakerestclient "k8s.io/client-go/rest/fake"
	"sigs.k8s.io/yaml"
)

const testdataWorkflow = "./testdata/workflow.yaml"

func TestRunnerStartHappyPath(t *testing.T) {
	t.Run("Should create Argo Workflow if not exits", func(t *testing.T) {
		// given
		input, expOutStatus := fixStartInputAndOutput(t)

		var expWf = struct {
			Spec wfv1.WorkflowSpec `json:"workflow"`
		}{}
		require.NoError(t, yaml.Unmarshal(input.Args, &expWf))

		fakeCli := fake.NewSimpleClientset()
		r := NewRunner(fakeCli)

		// when
		gotOutStatus, err := r.Start(context.Background(), input)

		// then
		require.NoError(t, err)

		require.NotNil(t, gotOutStatus)
		assert.Equal(t, expOutStatus, *gotOutStatus)

		gotWf, err := fakeCli.ArgoprojV1alpha1().Workflows(input.Ctx.Platform.Namespace).Get(input.Ctx.Name, metav1.GetOptions{})
		require.NoError(t, err)
		assert.EqualValues(t, expWf.Spec, gotWf.Spec)
	})

	t.Run("Should create a dry run request", func(t *testing.T) {
		// given
		input, expOutStatus := fixStartInputAndOutput(t)

		fakeCli := fake.NewSimpleClientset()
		assertDryRunReq := func(request *http.Request) (*http.Response, error) {
			_, dryRun := request.URL.Query()["dryRun"]
			assert.True(t, dryRun)

			return &http.Response{StatusCode: http.StatusOK, Body: emptyBody()}, nil
		}

		mockedRestCli := &WrapRESTClientset{fakeCli, assertDryRunReq}
		r := NewRunner(mockedRestCli)

		// when
		gotOutStatus, err := r.Start(context.Background(), input)

		// then
		require.NoError(t, err)

		require.NotNil(t, gotOutStatus)
		assert.Equal(t, expOutStatus, *gotOutStatus)
	})
}

func TestRunnerStartFailure(t *testing.T) {
	t.Run("Should return error when Argo Workflow already exits", func(t *testing.T) {
		// given
		input, _ := fixStartInputAndOutput(t)

		wf := fixFinishedArgoWorkflow(t, input.Ctx.Name, input.Ctx.Platform.Namespace)
		fakeCli := fake.NewSimpleClientset(&wf)

		r := NewRunner(fakeCli)

		// when
		out, err := r.Start(context.Background(), input)

		// then
		assert.EqualError(t, err, `while creating Argo Workflow: workflows.argoproj.io "Rocket" already exists`)
		assert.Nil(t, out)
	})

	t.Run("Should fail when manifest is malformed", func(t *testing.T) {
		// given
		input := runner.StartInput{
			Args: []byte("{malformed manifest"),
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
	t.Run("Should return success for successfully finished workflow", func(t *testing.T) {
		// given
		input := runner.WaitForCompletionInput{
			Ctx: runner.Context{
				Name: "Rocket",
				Platform: runner.KubernetesPlatformConfig{
					Namespace: "argo-ns",
				},
			},
		}

		wf := fixFinishedArgoWorkflow(t, input.Ctx.Name, input.Ctx.Platform.Namespace)

		fakeCli := fake.NewSimpleClientset(&wf)
		r := NewRunner(fakeCli)

		// when
		err := waitForFunc(func(ctx context.Context) error {
			out, err := r.WaitForCompletion(ctx, input)
			assert.True(t, out.Succeeded)
			return err
		}, 50*time.Millisecond)

		// then
		require.NoError(t, err)
	})

	t.Run("Should skip watch action for dry-run mode", func(t *testing.T) {
		// given
		input := runner.WaitForCompletionInput{
			Ctx: runner.Context{
				DryRun: true,
			},
		}

		r := NewRunner(nil)

		// when
		err := waitForFunc(func(ctx context.Context) error {
			out, err := r.WaitForCompletion(ctx, input)
			assert.True(t, out.Succeeded)
			return err
		}, 50*time.Millisecond)

		// then
		require.NoError(t, err)
	})
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

	var wf = struct {
		Spec wfv1.WorkflowSpec `json:"workflow"`
	}{}
	require.NoError(t, yaml.Unmarshal(rawWfSpec, &wf.Spec))

	return wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: wf.Spec,
		Status: wfv1.WorkflowStatus{
			FinishedAt: metav1.Now(),
			Phase:      wfv1.NodeSucceeded,
		},
	}
}

func fixStartInputAndOutput(t *testing.T) (runner.StartInput, runner.StartOutput) {
	inputManifest, err := ioutil.ReadFile(testdataWorkflow)
	require.NoError(t, err)

	input := runner.StartInput{
		Ctx: runner.Context{
			Name: "Rocket",
			Platform: runner.KubernetesPlatformConfig{
				Namespace: "argo-ns",
			},
		},
		Args: inputManifest,
	}

	expOutStatus := runner.StartOutput{
		Status: Status{
			ArgoWorkflowRef: WorkflowRef{
				Name:      input.Ctx.Name,
				Namespace: input.Ctx.Platform.Namespace,
			},
		},
	}

	return input, expOutStatus
}

type WrapRESTClientset struct {
	*fake.Clientset
	httpRoundTripper func(*http.Request) (*http.Response, error)
}

type FakeArgoprojV1alpha1RESTClient struct {
	argoprojv1alpha1.ArgoprojV1alpha1Interface
	httpRoundTripper func(*http.Request) (*http.Response, error)
}

func (c *FakeArgoprojV1alpha1RESTClient) RESTClient() rest.Interface {
	return &fakerestclient.RESTClient{
		NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		Client:               fakerestclient.CreateHTTPClient(c.httpRoundTripper),
	}
}

func (c *WrapRESTClientset) ArgoprojV1alpha1() argoprojv1alpha1.ArgoprojV1alpha1Interface {
	return &FakeArgoprojV1alpha1RESTClient{c.Clientset.ArgoprojV1alpha1(), c.httpRoundTripper}
}

func emptyBody() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte("{}")))
}
