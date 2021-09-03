package create

import (
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli/config"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// K3dDefaultNodeImage defines default Kubernetes image for a new k3d cluster.
	K3dDefaultNodeImage = "docker.io/rancher/k3s:v1.20.7-k3s1"
)

// K3dOptions holds configuration for creating k3d cluster.
type K3dOptions struct {
	Name string
	Wait time.Duration
}

// K3dDefaultConfig returns default set of values for k3d.
// We cannot use v1alpha2.SimpleConfig struct as tags are messed up and we are not able to marshal it properly.
var K3dDefaultConfig = fmt.Sprintf(`
kind: Simple
apiVersion: k3d.io/v1alpha2
name: %s
servers: 1
agents: 0
image: %s
network: %s
ports:
    - port: 80:80
      nodeFilters:
        - loadbalancer
    - port: 443:443
      nodeFilters:
        - loadbalancer
options:
    k3s:
        extraServerArgs:
            - --no-deploy=traefik
            - --node-label=ingress-ready=true
`, DefaultClusterName, K3dDefaultNodeImage, DefaultDockerNetwork)

// K3dSetDefaultConfig sets default values for k3d flags
func K3dSetDefaultConfig(flags *pflag.FlagSet) error {
	configFlag := flags.Lookup("config")
	if configFlag.Changed { // do not change user settings
		return nil
	}

	file, err := config.GetDefaultConfigPath("k3d-config.yaml")
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, []byte(K3dDefaultConfig), 0600); err != nil {
		return err
	}

	return configFlag.Value.Set(file)
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
