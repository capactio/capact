package helm

import (
	"context"
	"io/ioutil"
	"log"

	"helm.sh/helm/v3/pkg/engine"

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
}

func newHelmRunner(k8sCfg *rest.Config, cfg Config) *helmRunner {
	return &helmRunner{
		cfg:    cfg,
		k8sCfg: k8sCfg,
	}
}

func (r *helmRunner) Do(ctx context.Context, in runner.StartInput) (*runner.WaitForCompletionOutput, error) {
	namespace := in.ExecCtx.Platform.Namespace

	actionConfig, err := r.initActionConfig(namespace, log.Printf)
	if err != nil {
		return nil, err
	}

	cmdInput, cmdType, err := r.readCommandData(in)
	if err != nil {
		return nil, err
	}

	var helmCmd helmCommand
	switch cmdType {
	case InstallCommandType:
		renderer := newHelmRenderer(&engine.Engine{})
		helmCmd = newInstaller(r.cfg.RepositoryCachePath, actionConfig, renderer)
	default:
		return nil, errors.New("Unsupported command")
	}

	out, status, err := helmCmd.Do(ctx, cmdInput)
	if err != nil {
		return nil, errors.Wrapf(err, "while running Helm command '%s'", cmdType)
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
	return "helm"
}

func (r *helmRunner) initActionConfig(namespace string, debugLog action.DebugLog) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   &r.k8sCfg.Host,
		Insecure:    &r.k8sCfg.Insecure,
		CAFile:      &r.k8sCfg.CAFile,
		BearerToken: &r.k8sCfg.BearerToken,
	}

	err := actionConfig.Init(helmCfg, namespace, r.cfg.HelmDriver, debugLog)

	if err != nil {
		return nil, errors.Wrap(err, "while initializing Helm configuration")
	}

	return actionConfig, nil
}

func (r *helmRunner) readCommandData(in runner.StartInput) (Input, CommandType, error) {
	var args Arguments
	err := yaml.Unmarshal(in.Args, &args)
	if err != nil {
		return Input{}, "", errors.Wrap(err, "while unmarshaling runner arguments")
	}

	return Input{
		Args:    args,
		ExecCtx: in.ExecCtx,
	}, args.Command, nil
}

func (r *helmRunner) saveOutput(out Output) error {
	log.Printf("Saving default output to %s\n...", out.Default.Path)
	err := r.saveToFile(out.Default.Path, out.Default.Value)
	if err != nil {
		return errors.Wrap(err, "while saving default output")
	}

	if out.Additional == nil {
		return nil
	}

	log.Printf("Saving additional output to %s\n...", out.Additional.Path)
	err = r.saveToFile(out.Additional.Path, out.Additional.Value)
	if err != nil {
		return errors.Wrap(err, "while saving default output")
	}

	return nil
}

const defaultFilePermissions = 0644

func (r *helmRunner) saveToFile(path string, bytes []byte) error {
	err := ioutil.WriteFile(path, bytes, defaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to '%s'", path)
	}

	return nil
}
