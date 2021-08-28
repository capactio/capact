package delete

import (
	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/environment/delete"
	"capact.io/capact/internal/cli/printer"
	"fmt"
	"github.com/rancher/k3d/v4/cmd/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewK3d returns a cobra.Command for creating k3d environment.
func NewK3d() *cobra.Command {
	var opts delete.K3dOptions

	cmd := &cobra.Command{
		Use:   "k3d",
		Short: "Delete local k3d cluster",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Deleting cluster %q...", opts.Name))
			logrus.SetFormatter(spinnerFmt)
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return delete.K3d(cmd.Context(), opts)
		},
	}

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env delete kind --name capact-dev
	//   $ capact env delete k3d  --name capact-dev
	cmd.Flags().StringVar(&opts.Name, "name", create.DefaultClusterName, "Cluster name")
	_ = cmd.RegisterFlagCompletionFunc("name", util.ValidArgsAvailableClusters)

	cmd.Flags().BoolVar(&opts.RemoveRegistry, "remove-registry", true, "Remove registry")

	return cmd
}
