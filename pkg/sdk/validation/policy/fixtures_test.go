package policy_test

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixPolicyWithoutTypeRef() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{},
							Inject: &policy.InjectData{
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id",
										},
									},
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "id2",
											Description: ptr.String("ID 2"),
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id3",
											Name: "id-3",
										},
									},
								},
							},
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
						Path:     "cap.type.aws.*",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID: "31bb8355-10d7-49ce-a739-4554d8a40b63",
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.*",
						Revision: ptr.String("0.1.0"),
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

func fixPolicyWithWrongBackendForTypeRef() policy.Policy {
	return policy.Policy{
		TypeInstance: policy.TypeInstancePolicy{
			Rules: []policy.RulesForTypeInstance{
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							TypeRef: &types.TypeRef{
								Path:     "cap.type.hashicorp.vault.storage",
								Revision: "0.1.0",
							},
							ID:          "00fd161c-01bd-47a6-9872-47490e11f996",
							Description: ptr.String("Vault TI"),
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.type.aws.*",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							TypeRef: &types.TypeRef{
								Path:     "cap.type.aws.auth.secret-manager",
								Revision: "0.1.0",
							},
							ID: "31bb8355-10d7-49ce-a739-4554d8a40b63",
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.*",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							TypeRef: &types.TypeRef{
								Path:     "cap.core.type.hub.storage.postresql",
								Revision: "0.1.0",
							},
							ID:          "a36ed738-dfe7-45ec-acd1-8e44e8db893b",
							Description: ptr.String("Default Capact PostgreSQL backend"),
						},
					},
				},
			},
		},
	}
}

func fixPolicyWithTypeRef() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{},
							Inject: &policy.InjectData{
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id",
											TypeRef: &types.TypeRef{
												Path:     "cap.type.sample",
												Revision: "0.1.0",
											},
										},
									},
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "id2",
											Description: ptr.String("ID 2"),
											TypeRef: &types.TypeRef{
												Path:     "cap.type.sample2",
												Revision: "0.2.0",
											},
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id3",
											Name: "name",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.sample3",
											Revision: "0.3.0",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var implementationRevisionRaw = []byte(`
revision: 0.1.0
spec:
  additionalInput:
    parameters:
    - name: additional-parameters
      typeRef:
        path: cap.type.aws.auth.creds
        revision: 0.1.0
    - name: impl-specific-config
      typeRef:
        path: cap.type.aws.elasticsearch.install-input
        revision: 0.1.0
`)
