package delete

import (
	"context"

	"github.com/pkg/errors"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

type KindOptions struct {
	Name       string
	Kubeconfig string
}

func Kind(ctx context.Context, opts KindOptions) error {
	logger := cmd.NewLogger()
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
		cluster.ProviderWithDocker(),
	)
	// Delete individual cluster
	logger.V(0).Infof("Deleting cluster %q ...", opts.Name)
	if err := provider.Delete(opts.Name, opts.Kubeconfig); err != nil {
		return errors.Wrapf(err, "failed to delete cluster %q", opts.Name)
	}
	return nil
}
