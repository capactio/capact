package argo

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGCPGlobalPolicy() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
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
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "c268d3f5-8834-434b-bea2-b677793611c5",
										Description: ptr.String("GCP SA"),
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
	}
}

func fixAWSGlobalPolicy() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
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
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRefWithOptRevision{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
				},
			},
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.interface.aws.rds.postgresql.provision",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Requires: &[]types.ManifestRefWithOptRevision{
								{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRefWithOptRevision{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "517cf827-233c-4bf1-8fc9-48534424dd58",
										Description: ptr.String("GCP SA"),
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
	}
}

func fixGlobalPolicyForFallback() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Attributes: &[]types.ManifestRefWithOptRevision{
								{
									Path:     "cap.attribute.not-existing",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "517cf827-233c-4bf1-8fc9-48534424dd58",
										Description: ptr.String("GCP SA"),
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
	}
}

func fixTerraformPolicy() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.terraform.gcp.cloudsql.postgresql.install"),
						},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "c268d3f5-8834-434b-bea2-b677793611c5",
										Description: ptr.String("GCP SA"),
									},
								},
							},
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
	}
}
