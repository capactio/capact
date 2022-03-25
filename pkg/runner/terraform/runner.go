package terraform

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"capact.io/capact/internal/getter"

	"go.uber.org/zap"

	"capact.io/capact/pkg/runner"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

var _ runner.Runner = &terraformRunner{}

// Runner provides functionality to run and wait for Helm operations.
type terraformRunner struct {
	cfg Config
	log *zap.Logger

	terraform *terraform
}

// NewTerraformRunner returns a new Terraform runner instance.
func NewTerraformRunner(cfg Config) runner.Runner {
	return &terraformRunner{
		cfg: cfg,
	}
}

// Start starts the Terraform runner operation.
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

	// in case of git repository as a source, the module download needs to be done in empty directory
	r.log.Debug("Downloading source into workdir", zap.String("workdir", r.cfg.WorkDir))

	_, err = os.Stat(r.cfg.WorkDir)
	if strings.HasPrefix(args.Module.Source, "git") && !os.IsNotExist(err) {
		return nil, fmt.Errorf("the workdir directory %q must not exist when cloning git repository", r.cfg.WorkDir)
	}

	err = getter.Download(context.Background(), args.Module.Source, r.cfg.WorkDir, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while downloading module")
	}

	err = r.injectStateTypeInstance()
	if err != nil {
		return nil, errors.Wrap(err, "while splitting state TypeInstance")
	}

	err = r.mergeInputVariables(args.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "while merging variables")
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

// WaitForCompletion waits for the runner operation to complete.
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

	variables, err := r.terraform.variables()
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "while getting terraform.tfvars file")
	}

	state := StateTypeInstance{
		State:     tfstate,
		Variables: variables,
	}

	output := Output{
		Release:    release,
		Additional: additional,
		State:      state,
	}

	err = r.saveOutput(output)
	if err != nil {
		return &runner.WaitForCompletionOutput{}, errors.Wrap(err, "while saving output files")
	}

	return &runner.WaitForCompletionOutput{Succeeded: true, Message: "Terraform finished"}, nil
}

// Name returns the runner name.
func (r *terraformRunner) Name() string {
	return "terraform"
}

// InjectLogger sets the logger on the runner.
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

		if err := os.Setenv(k, v); err != nil {
			return errors.Wrapf(err, "while setting env %s", k)
		}
	}
	return nil
}

func (r *terraformRunner) injectStateTypeInstance() error {
	varsFilepath := path.Join(r.cfg.WorkDir, variablesFile)

	if r.cfg.StateTypeInstanceFilepath == "" {
		// no statefile was provided, create empty tfvars and return
		return runner.SaveToFile(varsFilepath, []byte("\n"))
	}

	data, err := ioutil.ReadFile(r.cfg.StateTypeInstanceFilepath)
	if err != nil {
		return errors.Wrapf(err, "while reading state file %s", r.cfg.StateTypeInstanceFilepath)
	}

	state := &StateTypeInstance{}
	if err := yaml.Unmarshal(data, state); err != nil {
		return errors.Wrap(err, "while unmarshaling StateTypeInstance")
	}

	stateFilepath := path.Join(r.cfg.WorkDir, stateFile)
	if err := runner.SaveToFile(stateFilepath, state.State); err != nil {
		return errors.Wrapf(err, "while writing state file %s", stateFilepath)
	}

	if err := runner.SaveToFile(varsFilepath, state.Variables); err != nil {
		return errors.Wrapf(err, "while writing vars file %s", varsFilepath)
	}

	return nil
}

func (r *terraformRunner) mergeInputVariables(variables string) error {
	// variables file has to end with a new line
	variables = variables + "\n"
	r.log.Debug("Writing Terraform variables file", zap.String("variables", variables), zap.String("workdir", r.cfg.WorkDir), zap.String("file", variablesFile))

	inputTfVarFilepath := "/tmp/input.tfvars"

	err := runner.SaveToFile(inputTfVarFilepath, []byte(variables))
	if err != nil {
		return errors.Wrap(err, "while saving input variables to file")
	}

	workdirTfVarsFilepath := path.Join(r.cfg.WorkDir, variablesFile)

	paths := []string{workdirTfVarsFilepath, inputTfVarFilepath}
	values, err := LoadVariablesFromFiles(paths...)
	if err != nil {
		return errors.Wrap(err, "while loading all variables files")
	}

	data := MarshalVariables(values)
	if err := runner.SaveToFile(workdirTfVarsFilepath, data); err != nil {
		return errors.Wrap(err, "while saving new variables to file")
	}

	return nil
}

func (r *terraformRunner) saveOutput(out Output) error {
	if out.Release != nil {
		r.log.Debug("Saving terraform release output", zap.String("path", r.cfg.Output.TerraformReleaseFilePath))
		bytesRelease, err := runner.NestingOutputUnderValue(out.Release)
		if err != nil {
			return errors.Wrap(err, "while nesting Terrafrom release under value")
		}

		err = runner.SaveToFile(r.cfg.Output.TerraformReleaseFilePath, bytesRelease)
		if err != nil {
			return errors.Wrap(err, "while saving terraform release output")
		}
	}

	if out.Additional != nil {
		r.log.Debug("Saving additional output", zap.String("path", r.cfg.Output.AdditionalFilePath))
		bytesAdditional, err := runner.NestingOutputUnderValue(out.Additional)
		if err != nil {
			return errors.Wrap(err, "while nesting Terrafrom additional under value")
		}

		err = runner.SaveToFile(r.cfg.Output.AdditionalFilePath, bytesAdditional)
		if err != nil {
			return errors.Wrap(err, "while saving default output")
		}
	}

	r.log.Debug("Saving state output", zap.String("path", r.cfg.Output.TfstateFilePath))
	stateData, err := yaml.Marshal(out.State)
	if err != nil {
		return errors.Wrap(err, "while marshaling state")
	}

	nestingStateData, err := runner.NestingOutputUnderValue(stateData)
	if err != nil {
		return errors.Wrap(err, "while nesting Terrafrom state data under value")
	}

	err = runner.SaveToFile(r.cfg.Output.TfstateFilePath, nestingStateData)
	if err != nil {
		return errors.Wrap(err, "while saving tfstate output")
	}

	return nil
}
