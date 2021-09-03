package delete

import (
	"context"
	"fmt"
	"os"
	"path"

	"capact.io/capact/internal/cli/printer"
	"github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	k3dtypes "github.com/rancher/k3d/v4/pkg/types"
	k3dutil "github.com/rancher/k3d/v4/pkg/util"
)

// K3d removes a given k3d cluster.
func K3d(ctx context.Context, name string) (err error) {
	status := printer.NewStatus(os.Stdout, "")
	defer func() {
		status.End(err == nil)
	}()

	c, err := client.ClusterGet(ctx, runtimes.SelectedRuntime, &k3dtypes.Cluster{Name: name})
	switch {
	case err == nil:
	case err == client.ClusterGetNoNodesFoundError:
		return nil
	default:
		return err
	}

	err = client.ClusterDelete(ctx, runtimes.SelectedRuntime, c, k3dtypes.ClusterDeleteOpts{SkipRegistryCheck: false})
	if err != nil {
		return err
	}

	status.Step("Removing cluster details from default kubeconfig...")
	if err := client.KubeconfigRemoveClusterFromDefaultConfig(ctx, c); err != nil {
		return err
	}
	status.Step("Removing standalone kubeconfig file (if there is one)...")
	configDir, err := k3dutil.GetConfigDirOrCreate()
	if err != nil {
		return err
	}

	kubeconfigfile := path.Join(configDir, fmt.Sprintf("kubeconfig-%s.yaml", c.Name))
	if err := os.Remove(kubeconfigfile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
