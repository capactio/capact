package statusreporter

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ConfigMapStatusEntryKey = "status"

var _ runner.StatusReporter = &K8sConfigMapReporter{}

// K8sConfigMapReporter provides functionality to report status from Action Runner in a way that K8s Engine can
// consume it later.
type K8sConfigMapReporter struct {
	cli client.Client
}

// NewK8sConfigMap returns new K8sConfigMapReporter instance.
func NewK8sConfigMap(cli client.Client) *K8sConfigMapReporter {
	return &K8sConfigMapReporter{
		cli: cli,
	}
}

// Report a given status to K8s Config Map, so K8s engine can consume it later.
func (c *K8sConfigMapReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	cm := &v1.ConfigMap{}
	key := client.ObjectKey{
		Name:      execCtx.Name,
		Namespace: execCtx.Platform.Namespace,
	}

	if err := c.cli.Get(ctx, key, cm); err != nil {
		return errors.Wrap(err, "while getting ConfigMap")
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	jsonStatus, err := json.Marshal(status)
	if err != nil {
		return errors.Wrap(err, "while marshaling status")
	}
	cm.Data[ConfigMapStatusEntryKey] = string(jsonStatus)

	if err := c.cli.Update(ctx, cm); err != nil {
		return errors.Wrap(err, "while updating ConfigMap")
	}

	return nil
}
