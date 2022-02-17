package policy

import (
	"testing"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
)

func TestToPolicy(t *testing.T) {
	// given
	givenWorkflowPolicy := workflowPolicyWithAdditionalInput(map[string]interface{}{"a": map[string]interface{}{"enabled": false}})
	expectedPolicy := policyWithAdditionalInput(map[string]interface{}{"a": map[string]interface{}{"enabled": false}})
	// when
	convertedPolicy, err := givenWorkflowPolicy.ToPolicy()
	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicy, convertedPolicy)
}

func workflowPolicyWithAdditionalInput(input map[string]interface{}) WorkflowPolicy {
	implementation := "cap.implementation.bitnami.postgresql.install"
	return WorkflowPolicy{
		Interface: WorkflowInterfacePolicy{
			Rules: WorkflowRulesList{
				WorkflowRulesForInterface{
					Interface: WorkflowInterfaceRef{
						ManifestRef: &types.ManifestRefWithOptRevision{
							Path: "cap.interface.database.postgresql.install",
						},
					},
					OneOf: []WorkflowRule{
						{
							ImplementationConstraints: ImplementationConstraints{
								Path: &implementation,
							},
							Inject: &WorkflowInjectData{
								AdditionalParameters: []AdditionalParametersToInject{
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

func policyWithAdditionalInput(input map[string]interface{}) Policy {
	implementation := "cap.implementation.bitnami.postgresql.install"
	return Policy{
		Interface: InterfacePolicy{
			Rules: InterfaceRulesList{
				RulesForInterface{
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.interface.database.postgresql.install",
					},
					OneOf: []Rule{
						{
							ImplementationConstraints: ImplementationConstraints{
								Path: &implementation,
							},
							Inject: &InjectData{
								AdditionalParameters: []AdditionalParametersToInject{
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
