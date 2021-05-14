package controller

import "time"

type Config struct {
	BuiltinRunner BuiltinRunnerConfig
}

type BuiltinRunnerConfig struct {
	Timeout time.Duration `envconfig:"default=30m"`
	Image   string
}
