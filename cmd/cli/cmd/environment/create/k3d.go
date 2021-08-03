package create

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/printer"

	"github.com/rancher/k3d/v4/cmd/cluster"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewK3D returns a cobra.Command for creating k3d environment.
func NewK3D() *cobra.Command {
	var name string

	k3d := cluster.NewCmdClusterCreate()
	k3d.Use = "k3d"
	k3d.Args = cobra.NoArgs
	k3d.Short = "Provision local k3d cluster"
	k3d.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Creating cluster %s...", name))
		logrus.SetFormatter(spinnerFmt)
		return create.K3dSetDefaultFlags(cmd.Flags())
	}
	k3d.RunE = func(cmd *cobra.Command, _ []string) (err error) {
		// Run k3d create cmd
		k3d.Run(cmd, []string{name})

		waitFlag := cmd.Flag("wait")
		if strings.EqualFold(waitFlag.Value.String(), "false") {
			return nil
		}

		rawTimeout := cmd.Flag("timeout").Value.String()
		timeout, err := time.ParseDuration(rawTimeout)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		return create.WaitForK3DReadyNodes(ctx, os.Stdout, name)
	}

	_ = k3d.Flags().Set("image", create.K3dDefaultNodeImage)

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env create kind --name capact-dev
	//   $ capact env create k3d  --name capact-dev
	k3d.Flags().StringVar(&name, "name", create.K3dDefaultClusterName, "Cluster name")

	return k3d
}
