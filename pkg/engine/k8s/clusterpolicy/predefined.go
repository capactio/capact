package clusterpolicy

import "capact.io/capact/pkg/sdk/apis/0.0.1/types"

// NewDenyAll returns a policy, which denies all Implementations.
func NewDenyAll() ClusterPolicy {
	return ClusterPolicy{
		APIVersion: CurrentAPIVersion,
		Rules:      nil,
	}
}

// NewAllowAll returns a policy, which allows all Implementations.
func NewAllowAll() ClusterPolicy {
	return ClusterPolicy{
		APIVersion: CurrentAPIVersion,
		Rules: RulesList{
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []Rule{
					{
						ImplementationConstraints: ImplementationConstraints{},
					},
				},
			},
		},
	}
}
