package argo

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/Project-Voltron/voltron/cmd/argo-runner/runner"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	k8sTesting "k8s.io/client-go/testing"
)

func Test(t *testing.T) {
	// given
	fakeCli := fake.NewSimpleClientset()
	r := NewRunner(fakeCli.ArgoprojV1alpha1())

	file, err := ioutil.ReadFile("./testdata/workflow.yaml")
	require.NoError(t, err)

	execCtx := runner.ExecutionContext{
		Name: "Rocket",
		Platform: runner.KubernetesPlatformConfig{
			Namespace: "argo-ns",
		},
	}

	// when
	err = r.Start(context.TODO(), execCtx, file)

	// then
	require.NoError(t, err)

	gotWf, err := fakeCli.ArgoprojV1alpha1().Workflows(execCtx.Platform.Namespace).Get(execCtx.Name, metav1.GetOptions{})
	require.NoError(t, err)

	spew.Dump(gotWf)
}

func assertT(t *testing.T) {
	t.Helper()

	//k8sTesting.NewCreateAction()

}

func checkAction(t *testing.T, expected, actual k8sTesting.Action) {
	t.Helper()

	assert.Truef(t, expected.Matches(actual.GetVerb(), actual.GetResource().Resource),
		"actions not matched: expected [%#v] got [%#v]", expected, actual)

	switch a := actual.(type) {
	case k8sTesting.CreateAction:
		e, ok := expected.(k8sTesting.CreateAction)
		assert.True(t, ok)

		expObject := e.GetObject()
		object := a.GetObject()

		assert.Equal(t, expObject, object)
	case k8sTesting.UpdateAction:
		e, ok := expected.(k8sTesting.UpdateAction)
		assert.True(t, ok)

		expObject := e.GetObject()
		object := a.GetObject()

		assert.Equal(t, expObject, object)
	case k8sTesting.PatchAction:
		e, ok := expected.(k8sTesting.PatchAction)
		assert.True(t, ok)

		expPatch := e.GetPatch()
		patch := a.GetPatch()

		assert.Equal(t, expPatch, patch)
	}
}
