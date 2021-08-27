package create

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rancher/k3d/v4/pkg/config/v1alpha2"
	"gopkg.in/yaml.v3"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// K3dDefaultNodeImage defines default Kubernetes image for a new k3d cluster.
	K3dDefaultNodeImage = "docker.io/rancher/k3s:v1.19.7-k3s1"
)

// K3dOptions holds configuration for creating k3d cluster.
type K3dOptions struct {
	Name string
	Wait time.Duration
}

// K3dDefaultConfig returns default set of values for k3d.
var K3dDefaultConfig = v1alpha2.SimpleConfig{
	TypeMeta: v1alpha2.TypeMeta{
		Kind:       "Simple",
		APIVersion: "k3d.io/v1alpha2",
	},
	Name:      DefaultClusterName,
	Servers:   1,
	Agents:    0,
	ExposeAPI: v1alpha2.SimpleExposureOpts{},
	Image:     K3dDefaultNodeImage,
	Network:   "capact",
	Ports: []v1alpha2.PortWithNodeFilters{
		{
			Port:        "80:80",
			NodeFilters: []string{"loadbalancer"},
		},
		{
			Port:        "443:443",
			NodeFilters: []string{"loadbalancer"},
		},
	},
	Options: v1alpha2.SimpleConfigOptions{
		K3sOptions: v1alpha2.SimpleConfigOptionsK3s{
			ExtraServerArgs: []string{"--no-deploy=traefik", "--node-label=ingress-ready=true"},
		},
	},
}

// K3dSetDefaultConfig sets default values for k3d flags
func K3dSetDefaultConfig(flags *pflag.FlagSet) error {
	config := flags.Lookup("config")
	if config.Changed { // do not change user settings
		return nil
	}

	file, err := os.CreateTemp("", "k3d-config")
	if err != nil {
		return err
	}
	out, err := yaml.Marshal(K3dDefaultConfig)
	if _, err := file.Write(out); err != nil {
		return errors.Wrap(err, "dup")
	}
	//defer os.Remove(file.Name())
	fmt.Println(file.Name())
	if err := config.Value.Set(file.Name()); err != nil {
		return err
	}

	return nil
}

// K3dRemoveWaitAndTimeoutFlags removes the wait and timeout flags
func K3dRemoveWaitAndTimeoutFlags(k3d *cobra.Command) {
	flags := k3d.Flags()
	k3d.ResetFlags()
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "wait" || flag.Name == "timeout" { // we set this by ourselves
			return
		}
		k3d.Flags().AddFlag(flag)
	})

}
