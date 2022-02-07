package policy_test

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGQLInput() graphql.PolicyInput {
	return graphql.PolicyInput{
		Interface: &graphql.InterfacePolicyInput{
			Rules: []*graphql.RulesForInterfaceInput{
				{
					Interface: &graphql.ManifestReferenceInput{
						Path:     "cap.interface.database.postgresql.install",
						Revision: ptr.String("0.1.0"),
					},
					OneOf: []*graphql.PolicyRuleInput{
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraintsInput{
								Requires: []*graphql.ManifestReferenceInput{
									{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
								Attributes: []*graphql.ManifestReferenceInput{
									{
										Path: "cap.attribute.cloud.provider.gcp",
									},
								},
							},
							Inject: &graphql.PolicyRuleInjectDataInput{
								RequiredTypeInstances: []*graphql.RequiredTypeInstanceReferenceInput{
									{
										ID:          "c268d3f5-8834-434b-bea2-b677793611c5",
										Description: ptr.String("Sample description"),
									},
								},
								AdditionalParameters: []*graphql.AdditionalParameterInput{
									{
										Name: "additional-parameters",
										Value: map[string]interface{}{
											"key1": "boom",
										},
									},
								},
								AdditionalTypeInstances: []*graphql.AdditionalTypeInstanceReferenceInput{
									{
										Name: "sample",
										ID:   "0b6dba9a-d111-419d-b236-357cf0e8603a",
									},
								},
							},
						},
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraintsInput{
								Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
							},
						},
					},
				},
				{
					Interface: &graphql.ManifestReferenceInput{
						Path: "cap.*",
					},
					OneOf: []*graphql.PolicyRuleInput{
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraintsInput{},
						},
					},
				},
			},
		},
		TypeInstance: &graphql.TypeInstancePolicyInput{
			Rules: []*graphql.RulesForTypeInstanceInput{
				{
					TypeRef: &graphql.ManifestReferenceInput{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: &graphql.TypeInstanceBackendInput{
						ID:          "00fd161c-01bd-47a6-9872-47490e11f996",
						Description: ptr.String("Vault TI"),
					},
				},
				{
					TypeRef: &graphql.ManifestReferenceInput{
						Path: "cap.type.aws.*",
					},
					Backend: &graphql.TypeInstanceBackendInput{
						ID: "31bb8355-10d7-49ce-a739-4554d8a40b63",
					},
				},
				{
					TypeRef: &graphql.ManifestReferenceInput{
						Path: "cap.*",
					},
					Backend: &graphql.TypeInstanceBackendInput{
						ID:          "a36ed738-dfe7-45ec-acd1-8e44e8db893b",
						Description: ptr.String("Default Capact PostgreSQL backend"),
					},
				},
			},
		},
	}
}

func fixGQL() graphql.Policy {
	return graphql.Policy{
		Interface: &graphql.InterfacePolicy{
			Rules: []*graphql.RulesForInterface{
				{
					Interface: &graphql.ManifestReferenceWithOptionalRevision{
						Path:     "cap.interface.database.postgresql.install",
						Revision: ptr.String("0.1.0"),
					},
					OneOf: []*graphql.PolicyRule{
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraints{
								Requires: []*graphql.ManifestReferenceWithOptionalRevision{
									{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
								Attributes: []*graphql.ManifestReferenceWithOptionalRevision{
									{
										Path: "cap.attribute.cloud.provider.gcp",
									},
								},
							},
							Inject: &graphql.PolicyRuleInjectData{
								RequiredTypeInstances: []*graphql.RequiredTypeInstanceReference{
									{
										ID:          "c268d3f5-8834-434b-bea2-b677793611c5",
										Description: ptr.String("Sample description"),
									},
								},
								AdditionalParameters: []*graphql.AdditionalParameter{
									{
										Name: "additional-parameters",
										Value: map[string]interface{}{
											"key1": "boom",
										},
									},
								},
								AdditionalTypeInstances: []*graphql.AdditionalTypeInstanceReference{
									{
										Name: "sample",
										ID:   "0b6dba9a-d111-419d-b236-357cf0e8603a",
									},
								},
							},
						},
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraints{
								Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
							},
						},
					},
				},
				{
					Interface: &graphql.ManifestReferenceWithOptionalRevision{
						Path: "cap.*",
					},
					OneOf: []*graphql.PolicyRule{
						{
							ImplementationConstraints: &graphql.PolicyRuleImplementationConstraints{},
						},
					},
				},
			},
		},
		TypeInstance: &graphql.TypeInstancePolicy{
			Rules: []*graphql.RulesForTypeInstance{
				{
					TypeRef: &graphql.ManifestReferenceWithOptionalRevision{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: &graphql.TypeInstanceBackend{
						ID:          "00fd161c-01bd-47a6-9872-47490e11f996",
						Description: ptr.String("Vault TI"),
					},
				},
				{
					TypeRef: &graphql.ManifestReferenceWithOptionalRevision{
						Path: "cap.type.aws.*",
					},
					Backend: &graphql.TypeInstanceBackend{
						ID: "31bb8355-10d7-49ce-a739-4554d8a40b63",
					},
				},
				{
					TypeRef: &graphql.ManifestReferenceWithOptionalRevision{
						Path: "cap.*",
					},
					Backend: &graphql.TypeInstanceBackend{
						ID:          "a36ed738-dfe7-45ec-acd1-8e44e8db893b",
						Description: ptr.String("Default Capact PostgreSQL backend"),
					},
				},
			},
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
								AdditionalParameters: []policy.AdditionalParametersToInject{
									{
										Name: "additional-parameters",
										Value: map[string]interface{}{
											"key1": "boom",
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											Name: "sample",
											ID:   "0b6dba9a-d111-419d-b236-357cf0e8603a",
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
		TypeInstance: policy.TypeInstancePolicy{
			Rules: []policy.RulesForTypeInstance{
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "00fd161c-01bd-47a6-9872-47490e11f996",
							Description: ptr.String("Vault TI"),
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.type.aws.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID: "31bb8355-10d7-49ce-a739-4554d8a40b63",
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "a36ed738-dfe7-45ec-acd1-8e44e8db893b",
							Description: ptr.String("Default Capact PostgreSQL backend"),
						},
					},
				},
			},
		},
	}
}
