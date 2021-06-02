package policy

import (
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fixCfgMap(t *testing.T, in policy.Policy) *v1.ConfigMap {
	policyStr, err := in.ToYAMLString()
	require.NoError(t, err)

	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyCfgMapName,
			Namespace: policyCfgMapNamespace,
		},
		Data: map[string]string{
			policyConfigMapKey: policyStr,
		},
	}
}

func fixModel() policy.Policy {
	return policy.Policy{
		APIVersion: policy.CurrentAPIVersion,
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Requires: &[]types.ManifestRef{
								{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRef{
								{
									Path: "cap.attribute.cloud.provider.gcp",
								},
							},
						},
						Inject: &policy.InjectData{
							TypeInstances: []policy.TypeInstanceToInject{
								{
									ID: "c268d3f5-8834-434b-bea2-b677793611c5",
									TypeRef: types.ManifestRef{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
							},
						},
					},
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}
