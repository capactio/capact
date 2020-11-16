package runner

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"io/ioutil"

)

type ActionRunner interface {
	Start(ctx context.Context, execCtx ExecutionContext, manifest []byte) error
	WaitForCompletion(ctx context.Context, execCtx ExecutionContext) error
}

// Runner provides generic runner service
type Runner struct {
	underlying ActionRunner
	cfg        Config
	log        *zap.Logger
}

func NewManager(underlying ActionRunner) (*Runner, error) {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuration")
	}

	// setup logger
	var logCfg zap.Config
	if cfg.LoggerDevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	log, err := logCfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "while creating zap logger")
	}

	return &Runner{
		underlying: underlying,
		cfg:        cfg,
		log:        log,
	}, nil
}

func (r *Runner) Execute(stop <-chan struct{}) error {
	ctx, cancel := r.cancelableContext(stop)
	defer cancel()

	manifest, err := ioutil.ReadFile(r.cfg.InputManifestPath)
	if err != nil {
		return errors.Wrap(err, "while reading manifest from disk")
	}

	err = r.underlying.Start(ctx, r.cfg.Context, manifest)
	if err != nil {
		return errors.Wrap(err, "while starting action")
	}

	err = r.underlying.WaitForCompletion(ctx, r.cfg.Context)
	if err != nil {
		return errors.Wrap(err, "while waiting for completion")
	}

	return nil
}

// cancelableContext returns context that is canceled when stop signal is received or configured timeout elapsed.
func (r *Runner) cancelableContext(stop <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	if r.cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.cfg.Timeout)
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-stop:
			cancel()
		}
	}()

	return ctx, cancel
}
