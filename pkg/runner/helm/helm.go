package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"projectvoltron.dev/voltron/internal/ptr"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

type helmCommand interface {
	Do(ctx context.Context, in Input) (Output, Status, error)
}

// Runner provides functionality to run and wait for Helm operations.
type helmRunner struct {
	cfg    Config
	k8sCfg *rest.Config
	log    *zap.Logger
}

func newHelmRunner(k8sCfg *rest.Config, cfg Config) *helmRunner {
	return &helmRunner{
		cfg:    cfg,
		k8sCfg: k8sCfg,
	}
}

func (r *helmRunner) Do(ctx context.Context, in runner.StartInput) (*runner.WaitForCompletionOutput, error) {
	namespace := in.RunnerCtx.Platform.Namespace

	actionConfig, err := r.initActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	cmdInput, err := r.readCommandData(in)
	if err != nil {
		return nil, err
	}

	renderer := NewRenderer()
	outputter := NewOutputter(r.log, renderer)

	var helmCmd helmCommand
	switch r.cfg.Command {
	case InstallCommandType:
		helmCmd = newInstaller(r.log, r.cfg.RepositoryCachePath, actionConfig, outputter)
	case UpgradeCommandType:
		helmCmd = newUpgrader(r.log, r.cfg.RepositoryCachePath, r.cfg.HelmReleasePath, actionConfig, outputter)
	default:
		return nil, errors.New("Unsupported command")
	}

	out, status, err := helmCmd.Do(ctx, cmdInput)
	if err != nil {
		return nil, errors.Wrapf(err, "while running Helm command %q", r.cfg.Command)
	}

	err = r.saveOutput(out)
	if err != nil {
		return nil, errors.Wrap(err, "while saving output")
	}

	return &runner.WaitForCompletionOutput{
		Succeeded: status.Succeeded,
		Message:   status.Message,
	}, nil
}

func (r *helmRunner) Name() string {
	return "helm.v3"
}

func (r *helmRunner) InjectLogger(logger *zap.Logger) {
	r.log = logger
}

func (r *helmRunner) initActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   &r.k8sCfg.Host,
		Insecure:    &r.k8sCfg.Insecure,
		CAFile:      &r.k8sCfg.CAFile,
		BearerToken: &r.k8sCfg.BearerToken,
		Namespace:   ptr.String(namespace),
	}

	debugLog := func(format string, v ...interface{}) {
		r.log.Debug(fmt.Sprintf(format, v...), zap.String("source", "Helm"))
	}

	err := actionConfig.Init(helmCfg, namespace, r.cfg.HelmDriver, debugLog)

	if err != nil {
		return nil, errors.Wrap(err, "while initializing Helm configuration")
	}

	return actionConfig, nil
}

func (r *helmRunner) readCommandData(in runner.StartInput) (Input, error) {
	var args Arguments
	err := yaml.Unmarshal(in.Args, &args)
	if err != nil {
		return Input{}, errors.Wrap(err, "while unmarshaling runner arguments")
	}

	return Input{
		Args: args,
		Ctx:  in.RunnerCtx,
	}, nil
}

func (r *helmRunner) saveOutput(out Output) error {
	r.log.Debug("Saving Helm release output", zap.String("path", r.cfg.Output.HelmReleaseFilePath))
	err := runner.SaveToFile(r.cfg.Output.HelmReleaseFilePath, out.Release)
	if err != nil {
		return errors.Wrap(err, "while saving Helm release output")
	}

	if out.Additional == nil {
		return nil
	}

	r.log.Debug("Saving additional output", zap.String("path", r.cfg.Output.AdditionalFilePath))
	err = runner.SaveToFile(r.cfg.Output.AdditionalFilePath, out.Additional)
	if err != nil {
		return errors.Wrap(err, "while saving default output")
	}

	return nil
}
