package create

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"capact.io/capact/internal/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"capact.io/capact/internal/cli/printer"

	"github.com/pkg/errors"
	"github.com/rancher/k3d/v4/pkg/client"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	k3dtypes "github.com/rancher/k3d/v4/pkg/types"
	"golang.org/x/sync/errgroup"
)

// K3dOptions holds configuration for creating k3d cluster.
type K3dOptions struct {
	Name            string
	Wait            time.Duration
	RegistryEnabled bool
	Registry        string
}

// WaitForK3dReadyNodes waits until nodes are in Ready state based on the role message log.
func WaitForK3dReadyNodes(ctx context.Context, w io.Writer, clusterName string) (err error) {
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

// K3dSetDefaultConfig sets default values for k3d flags
// We cannot use v1alpha2.SimpleConfig struct as tags are messed up and we are not able to marshal it properly.
func K3dSetDefaultConfig(flags *pflag.FlagSet, opts K3dOptions) error {
	configFlag := flags.Lookup("config")
	if configFlag.Changed { // do not change user settings
		return nil
	}

	fName, err := config.GetDefaultConfigPath("k3d-config.yaml")
	if err != nil {
		return err
	}

	tmpl, err := template.New("cfg").Parse(k3dDefaultConfigTmpl)
	if err != nil {
		return err
	}

	data := map[string]string{
		"name":        DefaultClusterName,
		"image":       k3dDefaultNodeImage,
		"networkName": K3dDockerNetwork,
	}
	if opts.RegistryEnabled {
		data["registry"] = ContainerRegistry
	}

	f, err := os.OpenFile(filepath.Clean(fName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	return configFlag.Value.Set(fName)
}

// K3dRemoveWaitAndTimeoutFlags removes the wait and timeout flags
func K3dRemoveWaitAndTimeoutFlags(k3d *cobra.Command) {
	flags := k3d.Flags()
	k3d.ResetFlags()
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "wait" || flag.Name == "timeout" { // we set this by ourselves
			return
		}
		if flag.Name == "volume" {
			flag.Shorthand = "" // to avoid 'unable to redefine 'v' shorthand in "k3d" flagset: it's already used for "volume" flag'
		}
		k3d.Flags().AddFlag(flag)
	})
}
