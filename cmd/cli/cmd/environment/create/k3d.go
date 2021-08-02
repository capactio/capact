package create

import (
	"context"
	"fmt"
	"github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	"golang.org/x/sync/errgroup"
	"time"

	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/printer"

	"github.com/rancher/k3d/v4/cmd/cluster"
	k3dtypes "github.com/rancher/k3d/v4/pkg/types"
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
	k3d.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		spinnerFmt := printer.NewLogrusSpinnerFormatter(fmt.Sprintf("Creating cluster %s ...", name))
		logrus.SetFormatter(spinnerFmt)

		create.K3dSetDefaultFlags(cmd.Flags())
	}
	k3d.RunE = func(cmd *cobra.Command, _ []string) error {
		k3d.Run(cmd, []string{name})
		return nil
	}
	k3d.PostRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Flag("wait").Value.String() == "true" {
			runtime := runtimes.Docker
			nodes, err := runtime.GetNodesByLabel(cmd.Context(), map[string]string{k3dtypes.LabelClusterName: name})
			if err != nil {
				logrus.Errorf("Failed to get nodes for cluster '%s'", name)
			}
			return NodeAddToClusterMulti(cmd.Context(), runtime, nodes)
		}
		return nil
	}

	k3d.Flags().Set("image", create.K3dDefaultNodeImage)

	// add `name` flag to have the same UX as we have for `kind` in the minimal scenario:
	//   $ capact env create kind --name capact-dev
	//   $ capact env create kind --name capact-dev
	k3d.Flags().StringVar(&name, "name", create.K3dDefaultClusterName, "Cluster name")

	return k3d
}

// NodeAddToClusterMulti adds multiple nodes to a chosen cluster
func NodeAddToClusterMulti(ctx context.Context, runtime runtimes.Runtime, nodes []*k3dtypes.Node) error {
	defer logrus.Trace("")
	nodeWaitGroup, ctx := errgroup.WithContext(ctx)
	for _, node := range nodes {
		currentNode := node
		nodeWaitGroup.Go(func() error {
			logrus.Infof("Wait for node '%s'", currentNode.Name)

			readyLogMessage := k3dtypes.ReadyLogMessageByRole[currentNode.Role]
			if currentNode.Role == k3dtypes.ServerRole {
				readyLogMessage = "Successfully registered node"
			}
			if readyLogMessage != "" {
				return client.NodeWaitForLogMessage(ctx, runtime, currentNode, readyLogMessage, time.Time{})
			}
			logrus.Warnf("NodeAddToClusterMulti: Set to wait for node %s to get ready, but there's no target log message defined", currentNode.Name)
			return nil
		})
	}
	if err := nodeWaitGroup.Wait(); err != nil {
		logrus.Errorln("Failed to bring up all nodes in time. Check the logs:")
		logrus.Errorf(">>> %+v", err)
		return fmt.Errorf("Failed to add nodes")
	}
	logrus.Info("All nodes are ready")

	return nil
}
