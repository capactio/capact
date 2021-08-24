package policy

import (
	"context"
	"io/ioutil"
	"testing"

	corev1alpha1 "capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	policyCfgMapName      = "policy-cfgmap"
	policyCfgMapNamespace = "policy-ns"
)

func TestService_Update(t *testing.T) {
	// given
	model := fixModel()
	cfgMap := fixCfgMap(t, model)

	svc, k8sCli := newServiceWithFakeClient(t, cfgMap)

	// change few properties in model
	model.Rules[0].Interface.Path = "cap.interface.updated.path"
	model.Rules[1].OneOf = []policy.Rule{
		{
			ImplementationConstraints: policy.ImplementationConstraints{
				Requires: &[]types.ManifestRefWithOptRevision{
					{
						Path: "cap.core.type.platform.kubernetes",
					},
				},
			},
		},
	}

	// when
	actual, err := svc.Update(context.Background(), model)

	// then
	require.NoError(t, err)

	assert.Equal(t, model, actual)
	getConfigMapAndAssertEqual(t, k8sCli, model)
}

func TestService_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		model := fixModel()
		cfgMap := fixCfgMap(t, model)

		svc, _ := newServiceWithFakeClient(t, cfgMap)

		// when
		actual, err := svc.Get(context.Background())

		// then
		require.NoError(t, err)
		assert.Equal(t, model, actual)
	})

	t.Run("Not found", func(t *testing.T) {
		// given
		svc, _ := newServiceWithFakeClient(t)

		// when
		_, err := svc.Get(context.Background())

		// then
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrPolicyConfigMapNotFound))
	})
}

func newServiceWithFakeClient(t *testing.T, objects ...runtime.Object) (*Service, client.Client) {
	k8sCli := fakeK8sClient(t, objects...)
	logger := zap.NewRaw(zap.UseDevMode(true), zap.WriteTo(ioutil.Discard))

	cfg := Config{
		Name:      policyCfgMapName,
		Namespace: policyCfgMapNamespace,
	}

	return NewService(logger, k8sCli, cfg), k8sCli
}

func fakeK8sClient(t *testing.T, objects ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	require.NoError(t, err)
	err = corev1alpha1.AddToScheme(scheme)
	require.NoError(t, err)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objects...).
		Build()
}

func getConfigMapAndAssertEqual(t *testing.T, k8sCli client.Client, expected policy.Policy) {
	var cfgMap v1.ConfigMap

	err := k8sCli.Get(context.Background(), client.ObjectKey{
		Name:      policyCfgMapName,
		Namespace: policyCfgMapNamespace,
	}, &cfgMap)
	assert.NoError(t, err)

	actual, err := policy.FromYAMLString(cfgMap.Data[policyConfigMapKey])
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
