package policy

// Config holds configuration for policy reference.
type Config struct {
	Name      string `envconfig:"default=capact-engine-cluster-policy"`
	Namespace string `envconfig:"default=capact-system"`
}
