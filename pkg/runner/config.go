package runner

import "time"

// Config holds whole configuration for Manager.
type Config struct {
	InputManifestPath string
	Context           ExecutionContext
	LoggerDevMode     bool          `envconfig:"default=false"`
	Timeout           time.Duration `envconfig:"optional"`
}

// ExecutionContext holds configuration directly connected with specific Action Runner.
type ExecutionContext struct {
	Name     string
	DryRun   bool `envconfig:"default=false"`
	Platform KubernetesPlatformConfig
}

// KubernetesPlatformConfig holds Kubernetes specific configuration can can be utilized by K8s Action Runners.
type KubernetesPlatformConfig struct {
	Namespace          string
	ServiceAccountName string
}
