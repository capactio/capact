package create

import (
	_ "embed"
)

//go:embed tmpl/k3d-config.tmpl.yaml
var k3dDefaultConfigTmpl string

const (
	// k3dDefaultNodeImage defines default Kubernetes image for a new k3d cluster.
	k3dDefaultNodeImage = "docker.io/rancher/k3s:v1.20.7-k3s1"
	// K3dDockerNetwork defines default Docker network name k3d cluster.
	K3dDockerNetwork = "capact"
)
