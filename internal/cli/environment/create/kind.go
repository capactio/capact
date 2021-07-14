package create

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

// KindOptions holds configuration for creating kind cluster.
type KindOptions struct {
	Name       string
	Config     string
	ImageName  string
	Retain     bool
	Wait       time.Duration
	Kubeconfig string
}

// Kind creates a new kind cluster.
func Kind(ctx context.Context, opts KindOptions) error {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
		cluster.ProviderWithDocker(),
	)

	options := []cluster.CreateOption{
		cluster.CreateWithNodeImage(opts.ImageName),
		cluster.CreateWithRetain(opts.Retain),
		cluster.CreateWithWaitForReady(opts.Wait),
		cluster.CreateWithKubeconfigPath(opts.Kubeconfig),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	}

	if opts.Config != "" {
		options = append(options, cluster.CreateWithConfigFile(opts.Config))
	} else {
		options = append(options, cluster.CreateWithV1Alpha4Config(KindDefaultConfig))
	}

	if err := provider.Create(
		opts.Name,
		options...,
	); err != nil {
		return errors.Wrap(err, "failed to provision cluster")
	}

	return nil
}
