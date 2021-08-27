package create

import (
	"context"
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/printer"

	"github.com/rancher/k3d/v4/cmd/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewK3d returns a cobra.Command for creating k3d environment.
func NewK3d() *cobra.Command {
	var opts create.K3dOptions

	k3d := cluster.NewCmdClusterCreate()
	k3d.Use = "k3d"
	k3d.Args = cobra.NoArgs
	k3d.Short = "Provision local k3d cluster"
	k3d.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Creating cluster %s...", opts.Name))
		logrus.SetFormatter(spinnerFmt)
		return create.K3dSetDefaultConfig(cmd.Flags(), opts)
	}
	k3d.RunE = func(cmd *cobra.Command, _ []string) (err error) {
		// Run k3d create cmd
		k3d.Run(cmd, []string{opts.Name})

		if opts.Wait == time.Duration(0) {
			return nil
		}
		ctx, cancel := getTimeoutContext(cmd.Context(), opts.Wait)
		defer cancel()
		return create.WaitForK3dReadyNodes(ctx, os.Stdout, opts.Name)
	}

	create.K3dRemoveWaitAndTimeoutFlags(k3d) // remove it so we use own `--wait` flag

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env create kind --name capact-dev --wait 10m
	//   $ capact env create k3d  --name capact-dev --wait 10m
	k3d.Flags().StringVar(&opts.Name, "name", create.DefaultClusterName, "Cluster name")
	k3d.Flags().DurationVar(&opts.Wait, "wait", time.Duration(0), "Wait for control plane node to be ready")
	k3d.Flags().BoolVar(&opts.RegistryEnabled, "registry", false, "Creates local registry and configures environment to use it")

	return k3d
}

func getTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout.Seconds() == 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
