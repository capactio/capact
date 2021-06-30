package renderer

import "time"

// Config stores the configuration for a Workflow renderer.
type Config struct {
	RenderTimeout time.Duration `envconfig:"default=10m"`
	MaxDepth      int           `envconfig:"default=50"`
}
