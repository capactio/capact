package renderer

import "time"

type Config struct {
	RenderTimeout   time.Duration `envconfig:"default=10m"`
	MaxDepth        int           `envconfig:"default=50"`
	OCHActionsImage string
}
