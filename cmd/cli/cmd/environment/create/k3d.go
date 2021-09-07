package create

import (
	"context"
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/printer"

	"github.com/rancher/k3d/v4/cmd/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewK3d returns a cobra.Command for creating k3d environment.
func NewK3d() *cobra.Command {
	var opts create.K3dOptions

	status := printer.NewStatus(os.Stdout, "")

	k3d := cluster.NewCmdClusterCreate()
	k3d.Use = "k3d"
	k3d.Args = cobra.NoArgs
	k3d.Short = "Provision local k3d cluster"
	// Needs to be `PersistentPreRunE` as the `PreRunE` is used by k3d to configure viper config.
	k3d.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Creating cluster %s...", opts.Name))
		logrus.SetFormatter(spinnerFmt)

		if err := create.K3dSetDefaultConfig(cmd.Flags(), opts); err != nil {
			return err
		}
		if !opts.RegistryEnabled {
			return nil
		}
		return create.LocalRegistry(cmd.Context(), status)
	}
	k3d.RunE = func(cmd *cobra.Command, _ []string) error {
		// 1. Create k3d cluster
		k3d.Run(cmd, []string{opts.Name})

		if opts.Wait == time.Duration(0) {
			return nil
		}
		// 2. Wait for k3d cluster
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Wait)
		defer cancel()
		return create.WaitForK3dReadyNodes(ctx, status, opts.Name)
	}
	k3d.PostRunE = func(cmd *cobra.Command, args []string) (err error) {
		if !opts.RegistryEnabled {
			return nil
		}
		if err := create.RegistryConnWithNetwork(cmd.Context(), create.K3dDockerNetwork); err != nil {
			return err
		}

		return capact.AddRegistryToHostsFile()
	}

	create.K3dRemoveWaitAndTimeoutFlags(k3d) // remove it, so we use own `--wait` flag

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env create kind --name capact-dev --wait 10m
	//   $ capact env create k3d  --name capact-dev --wait 10m
	k3d.Flags().StringVar(&opts.Name, "name", create.DefaultClusterName, "Cluster name")
	k3d.Flags().DurationVar(&opts.Wait, "wait", time.Duration(0), "Wait for control plane node to be ready")
	k3d.Flags().BoolVar(&opts.RegistryEnabled, "enable-registry", false, "Create Capact local Docker registry and configure k3d environment to use it")

	return k3d
}
