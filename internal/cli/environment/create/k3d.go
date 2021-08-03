package create

import (
	"context"
	"fmt"
	"io"
	"time"

	"capact.io/capact/internal/cli/printer"

	"github.com/pkg/errors"
	"github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	k3dtypes "github.com/rancher/k3d/v4/pkg/types"
	"golang.org/x/sync/errgroup"
)

// WaitForK3DReadyNodes waits until nodes are in Ready state based on the role message log.
func WaitForK3DReadyNodes(ctx context.Context, w io.Writer, clusterName string) (err error) {
	status := printer.NewStatus(w, "")
	defer func() {
		status.End(err == nil)
	}()

	status.Step("Waiting for '%s' cluster nodes to be ready", clusterName)

	runtime := runtimes.SelectedRuntime
	nodes, err := runtime.GetNodesByLabel(ctx, map[string]string{k3dtypes.LabelClusterName: clusterName})
	if err != nil {
		return fmt.Errorf("failed to get nodes for cluster %q", clusterName)
	}

	nodeWaitGroup, _ := errgroup.WithContext(ctx)
	for _, node := range nodes {
		currentNode := node
		nodeWaitGroup.Go(func() error {
			readyLogMessage := k3dtypes.ReadyLogMessageByRole[currentNode.Role]
			if currentNode.Role == k3dtypes.ServerRole {
				readyLogMessage = "Successfully initialized node"
			}
			if readyLogMessage != "" {
				return client.NodeWaitForLogMessage(ctx, runtime, currentNode, readyLogMessage, time.Time{})
			}
			return nil
		})
	}
	if err := nodeWaitGroup.Wait(); err != nil {
		return errors.Wrap(err, "while waiting for nodes to be up and ready")
	}

	return nil
}
