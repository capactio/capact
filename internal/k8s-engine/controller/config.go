package controller

import "time"

type Config struct {
	BuiltinRunner BuiltinRunnerConfig
	ClusterPolicy ClusterPolicyConfig
}

type BuiltinRunnerConfig struct {
	Timeout time.Duration `envconfig:"default=30m"`
	Image   string
}

type ClusterPolicyConfig struct {
	Name      string `envconfig:"default=capact-engine-cluster-policy"`
	Namespace string `envconfig:"default=capact-system"`
}
