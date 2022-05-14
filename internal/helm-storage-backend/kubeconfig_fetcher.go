package helmstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"

	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	helmrunner "capact.io/capact/pkg/runner/helm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

// TypeInstanceGetter is a interface for getting TypeInstance data.
type TypeInstanceGetter interface {
	FindTypeInstance(ctx context.Context, id string, opts ...local.TypeInstancesOption) (*hublocalgraphql.TypeInstance, error)
}

// KubeconfigFetcher is a type that gets kubeconfig from TypeInstance.
type KubeconfigFetcher struct {
	tiGetter TypeInstanceGetter
}

// NewKubeconfigFetcher returns a new KubeconfigFetcher object.
func NewKubeconfigFetcher(tiGetter TypeInstanceGetter) *KubeconfigFetcher {
	return &KubeconfigFetcher{tiGetter: tiGetter}
}

//FetchByTypeInstanceID returns kubeconfig TypeInstance data based on TypeInstance ID.
func (k *KubeconfigFetcher) FetchByTypeInstanceID(ctx context.Context, typeInstanceID string) (helmrunner.KubeconfigInput, error) {
	ti, err := k.tiGetter.FindTypeInstance(ctx, typeInstanceID, local.WithFields(local.TypeInstanceRootFields|local.TypeInstanceLatestResourceVersionValueField))
	if err != nil {
		return helmrunner.KubeconfigInput{}, errors.Wrapf(err, "while fetching TypeInstance %q", typeInstanceID)
	}

	if ti == nil {
		return helmrunner.KubeconfigInput{}, errors.Wrap(fmt.Errorf("TypeInstance with IO %q not found", typeInstanceID), "while getting TypeInstance")
	}

	if ti.LatestResourceVersion == nil || ti.LatestResourceVersion.Spec == nil {
		return helmrunner.KubeconfigInput{}, fmt.Errorf("invalid data fetched from Hub for TypeInstance %q", typeInstanceID)
	}

	valueBytes, err := json.Marshal(ti.LatestResourceVersion.Spec)
	if err != nil {
		return helmrunner.KubeconfigInput{}, errors.Wrapf(err, "while marshaling TypeInstance %q", typeInstanceID)
	}

	var kubeconfig helmrunner.KubeconfigInput
	err = json.Unmarshal(valueBytes, &kubeconfig)
	if err != nil {
		return helmrunner.KubeconfigInput{}, errors.Wrapf(err, "while unmarshalling TypeInstance %q into kubeconfig", typeInstanceID)
	}

	return kubeconfig, nil
}

// SetKubeconfigBasedOnTypeInstanceID sets a kubeconfig based on TypeInstance.
func (k *KubeconfigFetcher) SetKubeconfigBasedOnTypeInstanceID(ctx context.Context, logger *zap.Logger, typeInstanceID string) error {
	kubeconfig, err := k.FetchByTypeInstanceID(ctx, typeInstanceID)
	if err != nil {
		return errors.Wrap(err, "while getting kubeconfig TypeInstance")
	}
	kcfg, err := yaml.Marshal(kubeconfig.Value.Config)
	if err != nil {
		return errors.Wrap(err, "while marshaling kubeconfig TypeInstance")
	}
	return helmrunner.SetNewKubeconfig(kcfg, logger)
}
