package argo

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGCPClusterPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules: clusterpolicy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
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
						Inject: &clusterpolicy.InjectData{
							TypeInstances: []clusterpolicy.TypeInstanceToInject{
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
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}

func fixAWSClusterPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules: clusterpolicy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []clusterpolicy.Rule{

					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.ManifestRef{
								{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.interface.aws.rds.postgresql.provision",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.ManifestRef{
								{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
						Inject: &clusterpolicy.InjectData{
							TypeInstances: []clusterpolicy.TypeInstanceToInject{
								{
									ID: "517cf827-233c-4bf1-8fc9-48534424dd58",
									TypeRef: types.ManifestRef{
										Path:     "cap.type.aws.auth.credentials",
										Revision: ptr.String("0.1.0"),
									},
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}

func fixClusterPolicyForFallback() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules: clusterpolicy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Attributes: &[]types.ManifestRef{
								{
									Path:     "cap.attribute.not-existing",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
						Inject: &clusterpolicy.InjectData{
							TypeInstances: []clusterpolicy.TypeInstanceToInject{
								{
									ID: "gcp-sa-uuid",
									TypeRef: types.ManifestRef{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}

func fixTerraformPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules: clusterpolicy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.terraform.gcp.cloudsql.postgresql.install"),
						},
						Inject: &clusterpolicy.InjectData{
							TypeInstances: []clusterpolicy.TypeInstanceToInject{
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
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}
