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
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path:     "cap.interface.database.postgresql.install",
						Revision: ptr.String("0.1.0"),
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{
								Requires: &[]types.ManifestRefWithOptRevision{
									{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
								Attributes: &[]types.ManifestRefWithOptRevision{
									{
										Path: "cap.attribute.cloud.provider.gcp",
									},
								},
							},
							Inject: &policy.InjectData{
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "c268d3f5-8834-434b-bea2-b677793611c5",
											Description: ptr.String("Sample description"),
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
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{},
						},
					},
				},
			},
		},
	}
}
