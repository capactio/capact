package controller

import "time"

// Config holds Capact controller configuration.
type Config struct {
	BuiltinRunner BuiltinRunnerConfig
}

// BuiltinRunnerConfig holds configuration for built-in Action runner.
type BuiltinRunnerConfig struct {
	Timeout time.Duration `envconfig:"default=30m"`
	Image   string
}
