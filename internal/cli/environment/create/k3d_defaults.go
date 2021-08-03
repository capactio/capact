package create

import "github.com/spf13/pflag"

const (
	// K3dDefaultNodeImage defines default Kubernetes image for a new k3d cluster.
	K3dDefaultNodeImage = "docker.io/rancher/k3s:v1.19.7-k3s1"
	// K3dDefaultClusterName defines default name for a new k3d cluster.
	K3dDefaultClusterName = "dev-capact"
)

// Flags are used to set default values for k3d.
type Flags struct {
	Name   string
	Values []string
}

// K3dDefaultConfig returns default set of values for k3d.
var K3dDefaultConfig = []Flags{
	{
		Name:   "port",
		Values: []string{"80:80@loadbalancer", "443:443@loadbalancer"},
	},
	{
		Name:   "k3s-server-arg",
		Values: []string{"--no-deploy=traefik", "--node-label=ingress-ready=true"},
	},
	{
		Name:   "wait",
		Values: []string{"true"},
	},
	{
		Name:   "timeout",
		Values: []string{"60"},
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
