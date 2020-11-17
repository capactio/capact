package runner

import "time"

type Config struct {
	InputManifestPath string
	Context           ExecutionContext
	LoggerDevMode     bool          `envconfig:"default=false"`
	Timeout           time.Duration `envconfig:"optional"`
}

type ExecutionContext struct {
	Name     string
	DryRun   bool `envconfig:"default=false"`
	Platform KubernetesPlatformConfig
}

type KubernetesPlatformConfig struct {
	Namespace          string
	ServiceAccountName string
}
