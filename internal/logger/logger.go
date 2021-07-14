package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// New returns a new zap logger based on a given configuration.
func New(cfg Config) (*zap.Logger, error) {
	var logCfg zap.Config
	if cfg.DevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	logger, err := logCfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "while building zap logger")
	}

	return logger, nil
}
