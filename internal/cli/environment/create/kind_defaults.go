package create

import "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

const KubeadmKindConfigPatches = `kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    node-labels: "ingress-ready=true"
`
const (
	KindDefaultNodeImage   = "kindest/node:v1.19.1"
	KindDefaultClusterName = "kind-dev-capact"
)

var KindDefaultConfig = &v1alpha4.Cluster{
	TypeMeta: v1alpha4.TypeMeta{
		Kind:       "Cluster",
		APIVersion: "kind.x-k8s.io/v1alpha4",
	},
	Nodes: []v1alpha4.Node{
		{
			Role:                 v1alpha4.ControlPlaneRole,
			KubeadmConfigPatches: []string{KubeadmKindConfigPatches},
			ExtraPortMappings: []v1alpha4.PortMapping{
				{
					ContainerPort: 80,
					HostPort:      80,
					Protocol:      v1alpha4.PortMappingProtocolTCP,
				},
				{
					ContainerPort: 443,
					HostPort:      443,
					Protocol:      v1alpha4.PortMappingProtocolTCP,
				},
			},
		},
	},
}
