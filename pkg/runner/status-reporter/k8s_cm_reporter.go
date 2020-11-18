// Should be put in k8s-engine?
package statusreporter

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const cmStatusNameKey = "status"

type K8sConfigMapReporter struct {
	cli client.Client
}

func NewK8sConfigMap(cli client.Client) *K8sConfigMapReporter {
	return &K8sConfigMapReporter{
		cli: cli,
	}
}

func (c *K8sConfigMapReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	cm := &v1.ConfigMap{}
	err := c.cli.Get(ctx, client.ObjectKey{
		Name:      execCtx.Name,
		Namespace: execCtx.Platform.Namespace,
	}, cm)

	if err != nil {
		return errors.Wrap(err, "while getting ConfigMap")
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	jsonStatus, err := json.Marshal(status)
	if err != nil {
		return errors.Wrap(err, "while marshaling status")
	}
	cm.Data[cmStatusNameKey] = string(jsonStatus)

	err = c.cli.Update(ctx, cm)
	if err != nil {
		return errors.Wrap(err, "while updating ConfigMap")
	}

	return nil
}
