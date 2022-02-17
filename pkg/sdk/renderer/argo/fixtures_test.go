package argo

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGCPGlobalPolicy() policy.Policy {
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
		},
	}
}

func fixAWSGlobalPolicy(additionalParameters ...policy.AdditionalParametersToInject) policy.Policy {
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "517cf827-233c-4bf1-8fc9-48534424dd58",
											Description: ptr.String("AWS Credentials"),
										},
									},
								},
								AdditionalParameters: additionalParameters,
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

func fixGlobalPolicyForFallback() policy.Policy {
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "517cf827-233c-4bf1-8fc9-48534424dd58",
											Description: ptr.String("AWS Credentials"),
										},
									},
								},
							},
						},
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
							// No injects, even if the requirements are satisfied this Implementation should be ignored
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

func fixTerraformPolicy() policy.Policy {
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
								Path: ptr.String("cap.implementation.terraform.gcp.cloudsql.postgresql.install"),
							},
							Inject: &policy.InjectData{
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
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
		},
	}
}

func fixAWSRDSPolicy() policy.Policy {
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
								Path: ptr.String("cap.implementation.aws.rds.postgresql.install"),
							},
							Inject: &policy.InjectData{
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "517cf827-233c-4bf1-8fc9-48534424dd58",
											Description: ptr.String("AWS SA"),
										},
									},
								},
								AdditionalParameters: []policy.AdditionalParametersToInject{
									{
										Name: "additional-parameters",
										Value: map[string]interface{}{
											"region": "eu-central-1",
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
		},
	}
}

func fixExistingDBPolicy() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path:     "cap.interface.productivity.mattermost.install",
						Revision: ptr.String("0.1.0"),
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{},
							Inject: &policy.InjectData{
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											Name: "postgresql",
											ID:   "f2421415-b8a4-464b-be12-b617794411c5",
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
		},
	}
}
