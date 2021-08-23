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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"name": "capact",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"name": "capact",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"address": "1.2.3.4",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"address": "1.2.3.4",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"address": "1.2.3.4",
											"alias":   "karpatka",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123-111",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"name":  "capact",
											"alias": "capactio",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123-222",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
			expected: policy.Policy{
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"host": map[string]interface{}{
											"name":    "capact",
											"address": "1.2.3.4",
											"alias":   "karpatka",
										},
									},
									RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
										{
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "1314-142-123-111",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.gcp.auth.service-account",
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
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "action first then global for different interfaces - only rules",
			action: policy.ActionPolicy{
				Rules: policy.RulesList{
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
			global: policy.Policy{
				Rules: policy.RulesList{
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
			expected: policy.Policy{
				Rules: policy.RulesList{
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
			order: policy.MergeOrder{policy.Action, policy.Global},
		},
		{
			name: "merge type instances and additional input",
			action: policy.ActionPolicy{
				Rules: policy.RulesList{
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
									AdditionalInput: map[string]interface{}{
										"additional-input": map[string]interface{}{
											"a": 1,
										},
									},
								},
							},
						},
					},
				},
			},
			global: policy.Policy{
				Rules: policy.RulesList{
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
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "123-321-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.x",
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
				Rules: policy.RulesList{
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
											RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
												ID:          "123-321-123",
												Description: ptr.String("Sample TI"),
											},
											TypeRef: &types.ManifestRef{
												Path: "cap.type.x",
											},
										},
									},
									AdditionalInput: map[string]interface{}{
										"additional-input": map[string]interface{}{
											"a": 1,
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
			cli := client.NewPolicyEnforcedClient(nil)
			cli.SetPolicyOrder(tt.order)
			cli.SetGlobalPolicy(tt.global)
			cli.SetActionPolicy(tt.action)

			// expect
			assert.Equal(t, tt.expected, cli.Policy())
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

	cli := client.NewPolicyEnforcedClient(nil)

	err = cli.PushWorkflowStepPolicy(w1)
	assert.NoError(t, err)
	assert.Equal(t, expected1, cli.Policy())

	err = cli.PushWorkflowStepPolicy(w2)
	assert.NoError(t, err)
	assert.Equal(t, expected2, cli.Policy())

	cli.PopWorkflowStepPolicy()
	assert.Equal(t, expected1, cli.Policy())
}

func workflowPolicyWithAdditionalInput(input map[string]interface{}) policy.WorkflowPolicy {
	implementation := "cap.implementation.bitnami.postgresql.install"
	return policy.WorkflowPolicy{
		Rules: policy.WorkflowRulesList{
			policy.WorkflowRulesForInterface{
				Interface: policy.WorkflowInterfaceRef{
					ManifestRef: &types.ManifestRef{
						Path: "cap.interface.database.postgresql.install",
					},
				},
				OneOf: []policy.WorkflowRule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Path: &implementation,
						},
						Inject: &policy.WorkflowInjectData{
							AdditionalInput: input,
						},
					},
				},
			},
		},
	}
}
