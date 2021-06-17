package client_test

import (
	"testing"

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
		action   policy.Policy
		expected policy.Policy
		order    policy.MergeOrder
	}{
		{
			name: "only global policy",
			global: policy.Policy{
				Rules: policy.RulesList{
					policy.RulesForInterface{
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123",
											TypeRef: types.ManifestRef{
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
			action: policy.Policy{},
			expected: policy.Policy{
				Rules: policy.RulesList{
					policy.RulesForInterface{
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123",
											TypeRef: types.ManifestRef{
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
			action: policy.Policy{
				Rules: policy.RulesList{
					policy.RulesForInterface{
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123",
											TypeRef: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123",
											TypeRef: types.ManifestRef{
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
			action: policy.Policy{
				Rules: policy.RulesList{
					policy.RulesForInterface{
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123-111",
											TypeRef: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123-222",
											TypeRef: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
									TypeInstances: []policy.TypeInstanceToInject{
										{
											ID: "1314-142-123-111",
											TypeRef: types.ManifestRef{
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
			action: policy.Policy{
				Rules: policy.RulesList{
					policy.RulesForInterface{
						Interface: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
						Interface: types.ManifestRef{
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
