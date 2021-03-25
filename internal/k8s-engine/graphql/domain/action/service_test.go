package action_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/domain/action"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ptr"
	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestService_Create(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	svc, k8sCli := newServiceWithFakeClient(t)

	inputActionModel := fixModel(name)

	expected := inputActionModel.Action.DeepCopy()
	expected.Namespace = ns

	ctxWithNs := namespace.NewContext(context.Background(), ns)

	// when
	out, err := svc.Create(ctxWithNs, inputActionModel)

	// then
	require.NoError(t, err)

	assertActionEqual(t, *expected, out)
	findActionAndAssertEqual(t, k8sCli, *expected)
	getSecretAndAssertEqual(t, k8sCli, *inputActionModel.InputParamsSecret)
}

func TestService_Update(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	inputActionModel := fixModel(name)
	inputActionModel.SetNamespace(ns)

	svc, k8sCli := newServiceWithFakeClient(t, &inputActionModel.Action, inputActionModel.InputParamsSecret)

	ctxWithNs := namespace.NewContext(context.Background(), ns)

	// change few properties in model
	inputActionModel.Action.Spec.ActionRef = corev1alpha1.ManifestReference{
		Path:     "new.action",
		Revision: ptr.String("1.0.0"),
	}
	inputActionModel.InputParamsSecret.StringData = map[string]string{
		"parameters": `{"param":"new"}`,
	}

	// when
	out, err := svc.Update(ctxWithNs, inputActionModel)

	// then
	require.NoError(t, err)

	assertActionEqual(t, inputActionModel.Action, out)
	findActionAndAssertEqual(t, k8sCli, inputActionModel.Action)
	getSecretAndAssertEqual(t, k8sCli, *inputActionModel.InputParamsSecret)
}

func TestService_GetByName(t *testing.T) {
	const (
		name = "foo"
		ns   = "bar"
	)

	t.Run("Success", func(t *testing.T) {
		// given
		inputAction := fixModel(name).Action
		inputAction.Namespace = ns

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		actual, err := svc.GetByName(ctxWithNs, name)

		// then
		require.NoError(t, err)
		assertActionEqual(t, inputAction, actual)
	})

	t.Run("Not found", func(t *testing.T) {
		// given
		svc, _ := newServiceWithFakeClient(t)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		_, err := svc.GetByName(ctxWithNs, name)

		// then
		require.Error(t, err)
		assert.True(t, errors.Is(err, action.ErrActionNotFound))
	})
}

func TestService_List(t *testing.T) {
	// given
	const ns = "namespace"

	succeededPhase := corev1alpha1.SucceededActionPhase

	action1 := fixK8sActionMinimal("foo", ns, succeededPhase)
	action2 := fixK8sActionMinimal("bar", ns, succeededPhase)
	action3 := fixK8sActionMinimal("baz", ns, corev1alpha1.FailedActionPhase)

	testCases := []struct {
		Name   string
		Filter model.ActionFilter

		Expected []corev1alpha1.Action
	}{
		{
			Name:   "All items",
			Filter: model.ActionFilter{},
			Expected: []corev1alpha1.Action{
				action1, action2, action3,
			},
		},
		{
			Name:   "Filter by Succeeded phase",
			Filter: fixModelActionFilter(&succeededPhase),
			Expected: []corev1alpha1.Action{
				action1, action2,
			},
		},
	}

	//nolint:scopelint
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctxWithNs := namespace.NewContext(context.Background(), ns)
			svc, _ := newServiceWithFakeClient(t, &action1, &action2, &action3)

			// when
			actual, err := svc.List(ctxWithNs, testCase.Filter)

			// then
			require.NoError(t, err)
			assert.ElementsMatch(t, testCase.Expected, actual)
		})
	}
}

func TestService_DeleteByName(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	inputAction := fixModel(name).Action
	inputAction.Namespace = ns

	svc, k8sCli := newServiceWithFakeClient(t, &inputAction)
	ctxWithNs := namespace.NewContext(context.Background(), ns)

	// when
	err := svc.DeleteByName(ctxWithNs, name)

	// then
	require.NoError(t, err)

	var actual corev1alpha1.Action
	err = k8sCli.Get(context.Background(), client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, &actual)

	require.Error(t, err)
	assert.True(t, apierrors.IsNotFound(err))
}

func TestService_CancelByName(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	t.Run("Success", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.RunningActionPhase)
		inputAction.Spec.Run = ptr.Bool(true)

		svc, k8sCli := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.CancelByName(ctxWithNs, name)

		// then
		require.NoError(t, err)

		var actual corev1alpha1.Action
		err = k8sCli.Get(context.Background(), client.ObjectKey{
			Namespace: ns,
			Name:      name,
		}, &actual)
		require.NoError(t, err)
		require.NotNil(t, actual.Spec.Cancel)
		assert.True(t, *actual.Spec.Cancel)
	})

	t.Run("Error", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.InitialActionPhase)

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.CancelByName(ctxWithNs, name)

		// then
		require.Error(t, err)
		assert.True(t, errors.Is(err, action.ErrActionNotCancelable))
	})

	t.Run("Already Canceled", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.RunningActionPhase)
		inputAction.Spec.Cancel = ptr.Bool(true)

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.CancelByName(ctxWithNs, name)

		// then
		require.NoError(t, err)
	})
}

func TestService_RunByName(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	t.Run("Success", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.ReadyToRunActionPhase)

		svc, k8sCli := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.RunByName(ctxWithNs, name)

		// then
		require.NoError(t, err)

		var actual corev1alpha1.Action
		err = k8sCli.Get(context.Background(), client.ObjectKey{
			Namespace: ns,
			Name:      name,
		}, &actual)
		require.NoError(t, err)
		assert.NotNil(t, actual.Spec.Run)
		assert.True(t, *actual.Spec.Run)
	})

	t.Run("Error - Already Cancelled", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.InitialActionPhase)
		inputAction.Spec.Cancel = ptr.Bool(true)

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.RunByName(ctxWithNs, name)

		// then
		require.Error(t, err)
		assert.True(t, errors.Is(err, action.ErrActionCanceledNotRunnable))
	})

	t.Run("Error - Not ready to run", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.InitialActionPhase)
		inputAction.Status.Phase = corev1alpha1.BeingRenderedActionPhase

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.RunByName(ctxWithNs, name)

		// then
		require.Error(t, err)
		assert.True(t, errors.Is(err, action.ErrActionNotReadyToRun))
	})

	t.Run("Already Run", func(t *testing.T) {
		inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.RunningActionPhase)
		inputAction.Spec.Run = ptr.Bool(true)

		svc, _ := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.RunByName(ctxWithNs, name)

		// then
		require.NoError(t, err)
	})
}

func TestService_ContinueAdvancedRendering(t *testing.T) {
	// given
	const (
		name = "foo"
		ns   = "bar"
	)

	t.Run("Success - TypeInstances provided", func(t *testing.T) {
		inputAction := fixK8sActionForRenderingIteration(name, ns)
		renderingInput := model.AdvancedModeContinueRenderingInput{
			TypeInstances: &[]corev1alpha1.InputTypeInstance{
				{
					Name: "typeinstance1",
					ID:   "8349b98c-bb01-4ef0-a815-3b48454a3bd0",
				},
			},
		}

		svc, k8sCli := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.ContinueAdvancedRendering(ctxWithNs, name, renderingInput)

		// then
		require.NoError(t, err)

		var actual corev1alpha1.Action
		err = k8sCli.Get(context.Background(), client.ObjectKey{
			Namespace: ns,
			Name:      name,
		}, &actual)
		require.NoError(t, err)

		assert.Equal(t, actual.Status.Rendering.AdvancedRendering.RenderingIteration.CurrentIterationName, actual.Spec.AdvancedRendering.RenderingIteration.ApprovedIterationName)
		require.NotNil(t, actual.Spec.Input.TypeInstances)
		assert.Len(t, *actual.Spec.Input.TypeInstances, 2)

		expectedTypeInstance := (*renderingInput.TypeInstances)[0]

		var found bool
		for _, inputTypeInstance := range *actual.Spec.Input.TypeInstances {
			if inputTypeInstance.Name != expectedTypeInstance.Name {
				continue
			}

			assert.Equal(t, expectedTypeInstance, inputTypeInstance)
			found = true
			break
		}
		assert.True(t, found)
	})

	t.Run("Success - TypeInstances not provided", func(t *testing.T) {
		inputAction := fixK8sActionForRenderingIteration(name, ns)
		renderingInput := model.AdvancedModeContinueRenderingInput{}

		svc, k8sCli := newServiceWithFakeClient(t, &inputAction)

		ctxWithNs := namespace.NewContext(context.Background(), ns)

		// when
		err := svc.ContinueAdvancedRendering(ctxWithNs, name, renderingInput)

		// then
		require.NoError(t, err)

		var actual corev1alpha1.Action
		err = k8sCli.Get(context.Background(), client.ObjectKey{
			Namespace: ns,
			Name:      name,
		}, &actual)
		require.NoError(t, err)

		assert.Equal(t, actual.Status.Rendering.AdvancedRendering.RenderingIteration.CurrentIterationName, actual.Spec.AdvancedRendering.RenderingIteration.ApprovedIterationName)
	})

	t.Run("Error - invalid TypeInstances provided", func(t *testing.T) {
		// given
		const (
			name = "foo"
			ns   = "bar"
		)

		t.Run("Success", func(t *testing.T) {
			inputAction := fixK8sActionForRenderingIteration(name, ns)
			renderingInput := model.AdvancedModeContinueRenderingInput{
				TypeInstances: &[]corev1alpha1.InputTypeInstance{
					{
						Name: "invalid-name1",
						ID:   "8349b98c-bb01-4ef0-a815-3b48454a3bd0",
					},
					{
						Name: "invalid-name2",
						ID:   "8349b98c-bb01-4ef0-a815-3b48454a3bd0",
					},
				},
			}
			expectedErrMessage := "invalid set of TypeInstances provided for a given rendering iteration:"

			svc, _ := newServiceWithFakeClient(t, &inputAction)

			ctxWithNs := namespace.NewContext(context.Background(), ns)

			// when
			err := svc.ContinueAdvancedRendering(ctxWithNs, name, renderingInput)

			// then
			require.Error(t, err)
			assert.Contains(t, err.Error(), expectedErrMessage)
			for _, typeInstance := range *renderingInput.TypeInstances {
				assert.Contains(t, err.Error(), typeInstance.Name)
			}
		})

		t.Run("Error - Advanced rendering disabled", func(t *testing.T) {
			inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.ReadyToRunActionPhase)
			inputAction.Spec.AdvancedRendering = &corev1alpha1.AdvancedRendering{
				Enabled: false,
			}

			svc, _ := newServiceWithFakeClient(t, &inputAction)

			ctxWithNs := namespace.NewContext(context.Background(), ns)

			// when
			err := svc.ContinueAdvancedRendering(ctxWithNs, name, model.AdvancedModeContinueRenderingInput{})

			// then
			require.Error(t, err)
			assert.Error(t, err, action.ErrActionAdvancedRenderingDisabled)
		})

		t.Run("Error - Action not continuable", func(t *testing.T) {
			inputAction := fixK8sActionMinimal(name, ns, corev1alpha1.InitialActionPhase)
			inputAction.Spec.AdvancedRendering = &corev1alpha1.AdvancedRendering{
				Enabled: true,
			}

			svc, _ := newServiceWithFakeClient(t, &inputAction)

			ctxWithNs := namespace.NewContext(context.Background(), ns)

			// when
			err := svc.ContinueAdvancedRendering(ctxWithNs, name, model.AdvancedModeContinueRenderingInput{})

			// then
			require.Error(t, err)
			assert.True(t, errors.Is(err, action.ErrActionAdvancedRenderingIterationNotContinuable))
		})
	})
}

func findActionAndAssertEqual(t *testing.T, k8sCli client.Client, expected corev1alpha1.Action) {
	var actual corev1alpha1.Action

	err := k8sCli.Get(context.Background(), client.ObjectKey{
		Namespace: expected.Namespace,
		Name:      expected.Name,
	}, &actual)
	assert.NoError(t, err)

	assertActionEqual(t, expected, actual)
}

func assertActionEqual(t *testing.T, expected, actual corev1alpha1.Action) {
	actual.ResourceVersion = ""
	expected.ResourceVersion = ""
	assert.Equal(t, expected, actual)
}

func getSecretAndAssertEqual(t *testing.T, k8sCli client.Client, expected v1.Secret) {
	var actual v1.Secret

	err := k8sCli.Get(context.Background(), client.ObjectKey{
		Namespace: expected.Namespace,
		Name:      expected.Name,
	}, &actual)
	assert.NoError(t, err)

	actual.ResourceVersion = ""
	expected.ResourceVersion = ""

	assert.Equal(t, expected, actual)
}

func newServiceWithFakeClient(t *testing.T, objects ...runtime.Object) (*action.Service, client.Client) {
	k8sCli := fakeK8sClient(t, objects...)
	logger := zap.NewRaw(zap.UseDevMode(true), zap.WriteTo(ioutil.Discard))

	return action.NewService(logger, k8sCli), k8sCli
}

func fakeK8sClient(t *testing.T, objects ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	require.NoError(t, err)
	err = corev1alpha1.AddToScheme(scheme)
	require.NoError(t, err)

	return fake.NewFakeClientWithScheme(scheme, objects...)
}
