package clusterpolicy

var apiVersion = "0.1.0"

// DenyAll returns a policy, which denies all Implementations.
func NewDenyAll() ClusterPolicy {
	return ClusterPolicy{
		APIVersion: apiVersion,
		Rules:      nil,
	}
}

// AllowAll returns a policy, which allows all Implementations.
func NewAllowAll() ClusterPolicy {
	return ClusterPolicy{
		APIVersion: apiVersion,
		Rules: RulesMap{
			"cap.*": {
				OneOf: []Rule{
					{
						ImplementationConstraints: ImplementationConstraints{},
					},
				},
			},
		},
	}
}
