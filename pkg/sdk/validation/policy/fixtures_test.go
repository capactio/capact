package policy_test

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixPolicyWithoutTypeRef() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.RulesList{
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
										RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
											ID: "id",
										},
									},
									{
										RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
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
	}
}

func fixPolicyWithTypeRef() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.RulesList{
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
										RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
											ID: "id",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.sample",
											Revision: "0.1.0",
										},
									},
									{
										RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
											ID:          "id2",
											Description: ptr.String("ID 2"),
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.sample2",
											Revision: "0.2.0",
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
