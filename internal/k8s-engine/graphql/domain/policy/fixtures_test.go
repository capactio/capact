package policy_test

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGQLInput() graphql.PolicyInput {
	return graphql.PolicyInput{
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
							TypeInstances: []*graphql.TypeInstanceReferenceInput{
								{
									ID: "c268d3f5-8834-434b-bea2-b677793611c5",
									TypeRef: &graphql.ManifestReferenceInput{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
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
	}
}

func fixGQL() graphql.Policy {
	return graphql.Policy{
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
							TypeInstances: []*graphql.TypeInstanceReference{
								{
									ID: "c268d3f5-8834-434b-bea2-b677793611c5",
									TypeRef: &graphql.ManifestReferenceWithOptionalRevision{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
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
	}
}

func fixModel() clusterpolicy.ClusterPolicy {
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
