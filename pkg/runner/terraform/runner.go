package terraform

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

var _ runner.Runner = &terraformRunner{}

// Runner provides functionality to run and wait for Helm operations.
type terraformRunner struct {
	cfg Config
	log *zap.Logger

	terraform *terraform
}

func NewTerraformRunner(cfg Config) runner.Runner {
	return &terraformRunner{
		cfg: cfg,
	}
}

func (r *terraformRunner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	var args Arguments
	err := yaml.Unmarshal(in.Args, &args)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling runner arguments")
	}

	// both go-getter and terraform are using envs so setting them globally
	// it can be used to set credentials, paths to credentials, variables, args...
	err = r.setEnvVars(args.Env)
	if err != nil {
		return nil, errors.Wrap(err, "while proceeding Environment variables")
	}

	r.terraform = newTerraform(r.log, r.cfg.WorkDir, args)

	err = r.terraform.Start(in.RunnerCtx.DryRun)
	if err != nil {
		return nil, errors.Wrap(err, "while starting terraform")
	}

	return &runner.StartOutput{
		Status: "Running terraform",
	}, nil
}

func (r *terraformRunner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	if r.terraform == nil {
		return &runner.WaitForCompletionOutput{}, errors.New("terraform not started yet")
	}

	var cancel context.CancelFunc
	if in.RunnerCtx.Timeout.Duration() > 0 {
		ctx, cancel = context.WithTimeout(ctx, in.RunnerCtx.Timeout.Duration())
		defer cancel()
	}
	err := r.terraform.Wait(ctx)
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "terraform failed to finish")
	}

	release, err := r.terraform.releaseInfo()
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.New("failed to get release info")
	}
	additional, err := r.terraform.renderOutput()
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "while getting additional info")
	}
	tfstate, err := r.terraform.tfstate()
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "while getting terraform.tfstate file")
	}

	output := Output{
		Release:    release,
		Additional: additional,
		Tfstate:    tfstate,
	}

	err = r.saveOutput(output)
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "while saving output files")
	}

	return &runner.WaitForCompletionOutput{Succeeded: true, Message: "Terraform finished"}, nil
}

func (r *terraformRunner) Name() string {
	return "terraform"
}

func (r *terraformRunner) InjectLogger(logger *zap.Logger) {
	r.log = logger
}

func (r *terraformRunner) setEnvVars(env []string) error {
	for _, e := range env {
		s := strings.Split(e, "=")
		if len(s) < 2 {
			return fmt.Errorf("Invalid env variable %s", e)
		}
		k, v := s[0], s[1]
		os.Setenv(k, v)
	}
	return nil
}

func (r *terraformRunner) saveOutput(out Output) error {
	if out.Release != nil {
		r.log.Debug("Saving terraform release output", zap.String("path", r.cfg.Output.TerraformReleaseFilePath))
		err := runner.SaveToFile(r.cfg.Output.TerraformReleaseFilePath, out.Release)
		if err != nil {
			return errors.Wrap(err, "while saving terraform release output")
		}
	}

	if out.Additional != nil {
		r.log.Debug("Saving additional output", zap.String("path", r.cfg.Output.AdditionalFilePath))
		err := runner.SaveToFile(r.cfg.Output.AdditionalFilePath, out.Additional)
		if err != nil {
			return errors.Wrap(err, "while saving default output")
		}
	}

	if out.Tfstate != nil {
		r.log.Debug("Saving tfstate output", zap.String("path", r.cfg.Output.TfstateFilePath))
		err := runner.SaveToFile(r.cfg.Output.TfstateFilePath, out.Tfstate)
		if err != nil {
			return errors.Wrap(err, "while saving tfstate  output")
		}
	}

	return nil
}
