package delete

import (
	"fmt"

	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/environment/delete"
	"capact.io/capact/internal/cli/printer"
	"github.com/rancher/k3d/v4/cmd/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewK3D returns a cobra.Command for creating k3d environment.
func NewK3D() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "k3d",
		Short: "Delete local k3d cluster",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Deleting cluster %q...", name))
			logrus.SetFormatter(spinnerFmt)
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return delete.K3d(cmd.Context(), name)
		},
	}

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env delete kind --name capact-dev
	//   $ capact env delete k3d  --name capact-dev
	cmd.Flags().StringVar(&name, "name", create.K3dDefaultClusterName, "Cluster name")
	_ = cmd.RegisterFlagCompletionFunc("name", util.ValidArgsAvailableClusters)

	return cmd
}
