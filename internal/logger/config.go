package logger

// Config holds configuration for logger.
type Config struct {
	// DevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	DevMode bool `envconfig:"default=false"`
}
