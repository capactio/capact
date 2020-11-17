package runner

import (
	"context"
	"encoding/json"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

// Manager provides generic runner service
type Manager struct {
	runner         ActionRunner
	cfg            Config
	log            *zap.Logger
	statusReporter statusReporter
	k8sCli         client.Client
}

type statusReporter interface {
	Report(status interface{}) error
}

func NewManager(runner ActionRunner, cli client.Client) (*Manager, error) {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
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

	return &Manager{
		runner: runner,
		cfg:    cfg,
		log:    log,
		//statusReporter: statusReporter,
		k8sCli: cli,
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
	if err = r.ReportStatus(ctx, r.cfg.Context, out.Status); err != nil {
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

const cmStatusNameKey = "status"

func (r *Manager) ReportStatus(ctx context.Context, execCtx ExecutionContext, status interface{}) error {
	cm := &v1.ConfigMap{}
	err := r.k8sCli.Get(ctx, client.ObjectKey{
		Name:      execCtx.Name,
		Namespace: execCtx.Platform.Namespace,
	}, cm)

	if err != nil {
		return errors.Wrap(err, "while getting ConfigMap")
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	jsonStatus, err := json.Marshal(status)
	if err != nil {
		return errors.Wrap(err, "while marshaling status")
	}
	cm.Data[cmStatusNameKey] = string(jsonStatus)

	err = r.k8sCli.Update(ctx, cm)
	if err != nil {
		return errors.Wrap(err, "while updating ConfigMap")
	}

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
