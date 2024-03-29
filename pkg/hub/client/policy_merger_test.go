package client_test

import (
	"testing"

	"capact.io/capact/internal/ptr"

	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
)

func TestPolicyEnforcedClient_mergePolicies(t *testing.T) {
	interfacePath := "cap.interface.test.install"
	secondInterfacePath := "cap.interface.alibaba.install"
	implementationPath := "cap.implementation.test.install"
	secondImplementationPath := "cap.implementation.test.second.install"

	tests := []struct {
		name     string
		global   policy.Policy
		action   policy.ActionPolicy
		expected policy.Policy
		order    policy.MergeOrder
	}{
		{
			name: "only global policy",
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"name": "capact",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path:     "cap.type.gcp.auth.service-account",
														Revision: "0.1.0",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			action: policy.ActionPolicy{},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"name": "capact",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path:     "cap.type.gcp.auth.service-account",
														Revision: "0.1.0",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "only action policy",
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"address": "1.2.3.4",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			global: policy.Policy{},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"address": "1.2.3.4",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "action first then global for the same interface",
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"address": "1.2.3.4",
														"alias":   "karpatka",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123-111",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"name":  "capact",
														"alias": "capactio",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123-222",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional-global",
													Name: "additional-global",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &secondImplementationPath,
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"host": map[string]interface{}{
														"name":    "capact",
														"address": "1.2.3.4",
														"alias":   "karpatka",
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123-111",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &secondImplementationPath,
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "action first then global for different interfaces - only rules",
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: secondInterfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &secondImplementationPath,
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
								},
							},
						},
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: secondInterfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &secondImplementationPath,
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "merge type instances and additional input for default and rule",
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.a",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"additional-input": map[string]interface{}{
														"a": 1,
													},
												},
											},
										},
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123-111",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.b",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "123-321-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.x",
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional-global",
													Name: "additional-global",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample-global",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.a",
											Revision: "0.1.0",
										},
									},
								},
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.b",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
					Rules: policy.InterfaceRulesList{
						policy.RulesForInterface{
							Interface: types.ManifestRefWithOptRevision{
								Path: interfacePath,
							},
							OneOf: []policy.Rule{
								{
									ImplementationConstraints: policy.ImplementationConstraints{
										Path: &implementationPath,
									},
									Inject: &policy.InjectData{
										RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "123-321-123",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.x",
													},
												},
											},
											{
												TypeInstanceReference: policy.TypeInstanceReference{
													ID:          "1314-142-123-111",
													Description: ptr.String("Sample TI"),
													TypeRef: &types.TypeRef{
														Path: "cap.type.gcp.auth.service-account",
													},
												},
											},
										},
										AdditionalParameters: []policy.AdditionalParametersToInject{
											{
												Name: "additional-parameters",
												Value: map[string]interface{}{
													"additional-input": map[string]interface{}{
														"a": 1,
													},
												},
											},
										},
										AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional-global",
													Name: "additional-global",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample-global",
													Revision: "0.1.0",
												},
											},
											{
												AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
													ID:   "additional1",
													Name: "additional",
												},
												TypeRef: &types.ManifestRef{
													Path:     "cap.type.sample",
													Revision: "0.1.0",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			cli := client.NewPolicyEnforcedClient(nil, nil)
			cli.SetPolicyOrder(tt.order)
			cli.SetGlobalPolicy(tt.global)
			cli.SetActionPolicy(tt.action)

			// expect
			assert.Equal(t, tt.expected, cli.MergedPolicy())
		})
	}
}

func TestPolicyEnforcedClient_mergeTypeInstancePolicies(t *testing.T) {
	tests := []struct {
		name     string
		global   policy.Policy
		action   policy.ActionPolicy
		expected policy.Policy
		order    policy.MergeOrder
	}{
		{
			name: "only global policy",
			global: policy.Policy{
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
			},
			action: policy.ActionPolicy{},
			expected: policy.Policy{
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
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "only action policy",
			action: policy.ActionPolicy{
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
			},
			global: policy.Policy{},
			expected: policy.Policy{
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
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "action first then global for the same Types",
			action: policy.ActionPolicy{
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
			},
			global: policy.Policy{
				TypeInstance: policy.TypeInstancePolicy{
					Rules: []policy.RulesForTypeInstance{
						{
							TypeRef: types.ManifestRefWithOptRevision{
								Path:     "cap.type.aws.auth.credentials",
								Revision: ptr.String("0.1.0"),
							},
							Backend: policy.TypeInstanceBackend{
								TypeInstanceReference: policy.TypeInstanceReference{
									ID:          "1234-1234-1234-1234-1234",
									Description: ptr.String("Vault TI"),
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				TypeInstance: policy.TypeInstancePolicy{
					Rules: []policy.RulesForTypeInstance{
						{
							TypeRef: types.ManifestRefWithOptRevision{
								Path:     "cap.type.aws.auth.credentials",
								Revision: ptr.String("0.1.0"),
							},
							Backend: policy.TypeInstanceBackend{
								TypeInstanceReference: policy.TypeInstanceReference{
									ID:          "1234-1234-1234-1234-1234",
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
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "action first then global for different Types",
			action: policy.ActionPolicy{
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
			},
			global: policy.Policy{
				TypeInstance: policy.TypeInstancePolicy{
					Rules: []policy.RulesForTypeInstance{
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
					},
				},
			},
			expected: policy.Policy{
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
								Path: "cap.*",
							},
							Backend: policy.TypeInstanceBackend{
								TypeInstanceReference: policy.TypeInstanceReference{
									ID:          "a36ed738-dfe7-45ec-acd1-8e44e8db893b",
									Description: ptr.String("Default Capact PostgreSQL backend"),
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
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			cli := client.NewPolicyEnforcedClient(nil, nil)
			cli.SetPolicyOrder(tt.order)
			cli.SetGlobalPolicy(tt.global)
			cli.SetActionPolicy(tt.action)

			// expect
			assert.Equal(t, tt.expected, cli.MergedPolicy())
		})
	}
}

func TestPolicyEnforcedClient_mergeTypeInstancePoliciesForDefault(t *testing.T) {
	tests := []struct {
		name     string
		global   policy.Policy
		action   policy.ActionPolicy
		expected policy.Policy
		order    policy.MergeOrder
	}{
		{
			name: "TypeInstance with different Typeref for Global and Action Policy",
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("A"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "3333-3333-3333",
										Description: ptr.String("C"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.gcp.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("B"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.2.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("A"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "3333-3333-3333",
										Description: ptr.String("C"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.gcp.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("B"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.2.0",
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Global, policy.Action},
		},
		{
			name: "TypeInstance with the same Typeref for Global and Action Policy",
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("A"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("B"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("B"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Global, policy.Action},
		},
		{
			name: "TypeInstance with the same Typeref for Global and Action Policy but Action has higher priority",
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("A"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			action: policy.ActionPolicy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("B"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1234-123-123",
										Description: ptr.String("A"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.aws.auth.credentials",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			cli := client.NewPolicyEnforcedClient(nil, nil)
			cli.SetPolicyOrder(tt.order)
			cli.SetGlobalPolicy(tt.global)
			cli.SetActionPolicy(tt.action)

			// expect
			assert.Equal(t, tt.expected, cli.MergedPolicy())
		})
	}
}

func TestRequiredTypeInstancesForRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     policy.Rule
		global   policy.Policy
		expected []policy.RequiredTypeInstanceToInject
	}{
		{
			name: "Rule TypeInstance injection is preferred over Default",
			rule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "1111-2222-3333",
								Description: ptr.String("Sample TI"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: "0.1.0",
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1314-142-123",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.gcp.auth.service-account",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []policy.RequiredTypeInstanceToInject{
				{
					TypeInstanceReference: policy.TypeInstanceReference{
						ID:          "1111-2222-3333",
						Description: ptr.String("Sample TI"),
						TypeRef: &types.TypeRef{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.1.0",
						},
					},
				},
			},
		},
		{
			name: "TypeInstaces should not contain duplicates",
			rule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "1111-2222-3333",
								Description: ptr.String("Sample TI"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: "0.1.0",
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "1111-2222-3333",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.gcp.auth.service-account",
											Revision: "0.1.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []policy.RequiredTypeInstanceToInject{
				{
					TypeInstanceReference: policy.TypeInstanceReference{
						ID:          "1111-2222-3333",
						Description: ptr.String("Sample TI"),
						TypeRef: &types.TypeRef{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.1.0",
						},
					},
				},
			},
		},
		{
			name: "TypeInstaces are correctly distinguished based on Typeref",
			rule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "1111-1111-1111",
								Description: ptr.String("Sample TI"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: "0.1.0",
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Interface: policy.InterfacePolicy{
					Default: &policy.InterfaceDefault{
						Inject: &policy.DefaultInject{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									TypeInstanceReference: policy.TypeInstanceReference{
										ID:          "2222-2222-2222",
										Description: ptr.String("Sample TI"),
										TypeRef: &types.TypeRef{
											Path:     "cap.type.gcp.auth.service-account",
											Revision: "0.2.0",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []policy.RequiredTypeInstanceToInject{
				{
					TypeInstanceReference: policy.TypeInstanceReference{
						ID:          "2222-2222-2222",
						Description: ptr.String("Sample TI"),
						TypeRef: &types.TypeRef{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.2.0",
						},
					},
				},
				{
					TypeInstanceReference: policy.TypeInstanceReference{
						ID:          "1111-1111-1111",
						Description: ptr.String("Sample TI"),
						TypeRef: &types.TypeRef{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.1.0",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			cli := client.NewPolicyEnforcedClient(nil, nil)
			cli.SetGlobalPolicy(tt.global)

			// expect
			assert.Equal(t, tt.expected, cli.MergeRequiredTypeInstancesForRule(tt.rule))
		})
	}
}

func TestNestedWorkflowPolicy(t *testing.T) {
	w1 := workflowPolicyWithAdditionalInput(map[string]interface{}{"a": 1})
	w2 := workflowPolicyWithAdditionalInput(map[string]interface{}{"a": 2, "b": 3})

	expected1, err := workflowPolicyWithAdditionalInput(map[string]interface{}{"a": 1}).ToPolicy()
	assert.NoError(t, err)

	expected2, err := workflowPolicyWithAdditionalInput(map[string]interface{}{"a": 1, "b": 3}).ToPolicy()
	assert.NoError(t, err)

	cli := client.NewPolicyEnforcedClient(nil, nil)

	err = cli.PushWorkflowStepPolicy(w1)
	assert.NoError(t, err)
	assert.Equal(t, expected1, cli.MergedPolicy())

	err = cli.PushWorkflowStepPolicy(w2)
	assert.NoError(t, err)
	assert.Equal(t, expected2, cli.MergedPolicy())

	cli.PopWorkflowStepPolicy()
	assert.Equal(t, expected1, cli.MergedPolicy())
}

func workflowPolicyWithAdditionalInput(input map[string]interface{}) policy.WorkflowPolicy {
	implementation := "cap.implementation.bitnami.postgresql.install"
	return policy.WorkflowPolicy{
		Interface: policy.WorkflowInterfacePolicy{
			Rules: policy.WorkflowRulesList{
				policy.WorkflowRulesForInterface{
					Interface: policy.WorkflowInterfaceRef{
						ManifestRef: &types.ManifestRefWithOptRevision{
							Path: "cap.interface.database.postgresql.install",
						},
					},
					OneOf: []policy.WorkflowRule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{
								Path: &implementation,
							},
							Inject: &policy.WorkflowInjectData{
								AdditionalParameters: []policy.AdditionalParametersToInject{
									{
										Name:  "additional-parameters",
										Value: input,
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
