package policy

import "capact.io/capact/pkg/sdk/apis/0.0.1/types"

// NewDenyAll returns a policy, which denies all Implementations.
func NewDenyAll() Policy {
	return Policy{
		Interface: InterfacePolicy{
			Rules: nil,
		},
	}
}

// NewAllowAll returns a policy, which allows all Implementations.
func NewAllowAll() Policy {
	return Policy{
		Interface: InterfacePolicy{
			Rules: InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					OneOf: []Rule{
						{
							ImplementationConstraints: ImplementationConstraints{},
						},
					},
				},
			},
		},
	}
}
