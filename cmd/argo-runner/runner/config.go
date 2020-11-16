package runner

import "time"

type Config struct {
	InputManifestPath string
	Context           ExecutionContext
	LoggerDevMode     bool `envconfig:"default=false"`
	Timeout           time.Duration `envconfig:"optional"`
}

type ExecutionContext struct {
	Name     string
	Platform KubernetesPlatformConfig
}

type KubernetesPlatformConfig struct {
	Namespace string
}
