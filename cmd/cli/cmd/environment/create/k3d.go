package create

import (
	"capact.io/capact/internal/cli/environment/create"
	"github.com/rancher/k3d/v4/cmd/cluster"
	"github.com/spf13/cobra"
)

// NewK3D returns a cobra.Command for creating k3d environment.
func NewK3D() *cobra.Command {
	var name string

	k3d := cluster.NewCmdClusterCreate()
	k3d.Use = "k3d"
	k3d.Args = cobra.NoArgs
	k3d.Short = "Provision local k3d cluster"
	k3d.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		create.K3dSetDefaultFlags(k3d.Flags())
	}
	k3d.RunE = func(cmd *cobra.Command, args []string) error {
		k3d.Run(cmd, []string{name})
		return nil
	}

	k3d.Flags().Set("image", create.K3dDefaultNodeImage)

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env create kind --name capact-dev
	//   $ capact env create kind --name capact-dev
	k3d.Flags().StringVar(&name, "name", create.K3dDefaultClusterName, "Cluster name")

	return k3d
}
