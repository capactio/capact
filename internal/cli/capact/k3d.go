package capact

import (
	"capact.io/capact/internal/cli/printer"
	"context"
	"github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	"github.com/rancher/k3d/v4/pkg/tools"
	k3dtypes "github.com/rancher/k3d/v4/pkg/types"
	"github.com/sirupsen/logrus"
)

// LoadK3dImages loads Docker images into K3d environment
func LoadK3dImages(ctx context.Context, clusterName string, images []string) error {
	logrus.SetFormatter(printer.NewLogrusSpinnerFormatter(""))
	cluster, err := client.ClusterGet(ctx, runtimes.SelectedRuntime, &k3dtypes.Cluster{Name: clusterName})
	if err != nil {
		return err
	}

	err = tools.ImageImportIntoClusterMulti(ctx, runtimes.SelectedRuntime, images, cluster, k3dtypes.ImageImportOpts{})
	if err != nil {
		return err
	}
	return nil
}
