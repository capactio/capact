package create

import (
	"capact.io/capact/internal/cli/environment/create"
	"github.com/rancher/k3d/v4/cmd/cluster"
	"github.com/spf13/cobra"
)

// NewK3D returns a cobra.Command for creating k3d environment.
func NewK3D() *cobra.Command {
	var opts K3dOptions
	k3d := cluster.NewCmdClusterCreate()
	cmd := &cobra.Command{
		Use:   "k3d",
		Short: "Provision local k3d cluster",
		Long:  k3d.Long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			//
			k3d.Run(cmd, []string{opts.Name})
			return nil
		},
	}
	*cmd.Flags() = *k3d.Flags()

	cmd.Flags().StringVar(&opts.Name, "name", create.K3dDefaultClusterName, "Cluster name")
	cmd.Flag("image").Value.Set(create.K3dDefaultNodeImage)

	return cmd
}

// K3dOptions holds configuration for creating k3d cluster.
type K3dOptions struct {
	Name string
}
