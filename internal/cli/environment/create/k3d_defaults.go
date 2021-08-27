package create

import (
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

type vals []string

// Flags are used to set default values for k3d.
type Flags struct {
	Name   string
	Values vals
}

// K3dDefaultConfig returns default set of values for k3d.
var K3dDefaultConfig = []Flags{
	{
		Name:   "port",
		Values: vals{"80:80@loadbalancer", "443:443@loadbalancer"},
	},
	{
		Name:   "k3s-server-arg",
		Values: vals{"--no-deploy=traefik", "--node-label=ingress-ready=true"},
	},
	{
		Name:   "image",
		Values: vals{K3dDefaultNodeImage},
	},

}

// K3dSetDefaultFlags sets default values for k3d flags
func K3dSetDefaultFlags(flags *pflag.FlagSet) error {
	for _, cfg := range K3dDefaultConfig {
		flag := flags.Lookup(cfg.Name)
		if flag.Changed { // do not change user settings
			continue
		}
		for _, val := range cfg.Values {
			if err := flag.Value.Set(val); err != nil {
				return err
			}
		}
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
