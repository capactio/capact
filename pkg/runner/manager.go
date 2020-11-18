package runner

import (
	"context"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

// Manager provides generic runner service
type Manager struct {
	runner         ActionRunner
	cfg            Config
	log            *zap.Logger
	statusReporter StatusReporter
}

func NewManager(runner ActionRunner, statusReporter StatusReporter) (*Manager, error) {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuration")
	}

	log, err := getLogger(cfg.LoggerDevMode)
	if err != nil {
		return nil, errors.Wrap(err, "while creating zap logger")
	}

	loggerInto(log, runner)

	return &Manager{
		runner:         runner,
		cfg:            cfg,
		log:            log,
		statusReporter: statusReporter,
	}, nil
}

func (r *Manager) Execute(stop <-chan struct{}) error {
	log := r.log.With(zap.String("runner", r.runner.Name()))

	ctx, cancel := r.cancelableContext(stop)
	defer cancel()

	manifest, err := ioutil.ReadFile(r.cfg.InputManifestPath)
	if err != nil {
		return errors.Wrap(err, "while reading manifest from disk")
	}

	r.log.Debug("Starting runner")
	out, err := r.runner.Start(ctx, StartInput{
		ExecCtx:  r.cfg.Context,
		Manifest: manifest,
	})
	if err != nil {
		return errors.Wrap(err, "while starting action")
	}
	r.log.Debug("Runner started", zap.Any("status", out.Status))

	// save to disk or directly to CM and get rid of sidecar:
	// problem with flat workflow and we cannot add sidecar to a given step
	if err = r.statusReporter.Report(ctx, r.cfg.Context, out.Status); err != nil {
		return errors.Wrap(err, "while setting status")
	}

	log.Debug("Waiting for runner completion")
	err = r.runner.WaitForCompletion(ctx, WaitForCompletionInput{ExecCtx: r.cfg.Context})
	if err != nil {
		log.Error("while waiting for runner completion", zap.Error(err))
		return errors.Wrap(err, "while waiting for completion")
	}
	log.Debug("Manager job completed")

	return nil
}

// cancelableContext returns context that is canceled when stop signal is received or configured timeout elapsed.
func (r *Manager) cancelableContext(stop <-chan struct{}) (context.Context, context.CancelFunc) {
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

// LoggerInjector is used by the Manager to inject logger to Runner.
type LoggerInjector interface {
	InjectLogger(*zap.Logger)
}

// loggerInto will set logger on `runner` if requested.
func loggerInto(log *zap.Logger, runner interface{}) {
	if s, ok := runner.(LoggerInjector); ok {
		s.InjectLogger(log.Named("runner"))
	}
}

func getLogger(loggerDevMode bool) (*zap.Logger, error) {
	if loggerDevMode {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
