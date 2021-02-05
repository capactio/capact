package statusreporter

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const SecretStatusEntryKey = "status"

var _ runner.StatusReporter = &K8sSecretReporter{}

// K8sSecretReporter provides functionality to report status from Action Runner in a way that K8s Engine can
// consume it later.
type K8sSecretReporter struct {
	cli client.Client
}

// NewK8sSecret returns new K8sSecretReporter instance.
func NewK8sSecret(cli client.Client) *K8sSecretReporter {
	return &K8sSecretReporter{
		cli: cli,
	}
}

// Report a given status to K8s Secret, so K8s engine can consume it later.
func (c *K8sSecretReporter) Report(ctx context.Context, runnerCtx runner.Context, status interface{}) error {
	secret := &v1.Secret{}
	key := client.ObjectKey{
		Name:      runnerCtx.Name,
		Namespace: runnerCtx.Platform.Namespace,
	}

	if err := c.cli.Get(ctx, key, secret); err != nil {
		return errors.Wrap(err, "while getting Secret")
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	jsonStatus, err := json.Marshal(status)
	if err != nil {
		return errors.Wrap(err, "while marshaling status")
	}
	secret.Data[SecretStatusEntryKey] = jsonStatus

	if err := c.cli.Update(ctx, secret); err != nil {
		return errors.Wrap(err, "while updating Secret")
	}

	return nil
}
