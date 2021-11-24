package gitlabapi

import (
	"bytes"
	"context"
	"text/template"

	"capact.io/capact/pkg/runner"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

// RESTRunner provides functionality to execute REST API calls.
type RESTRunner struct {
	cfg Config
	log *zap.Logger
}

// Do sends an API request and renders the API response.
func (r *RESTRunner) Do(ctx context.Context, in runner.StartInput) (*runner.WaitForCompletionOutput, error) {
	input, err := r.readInputData(in)
	if err != nil {
		return nil, errors.Wrap(err, "while reading input data")
	}

	var opts []gitlab.ClientOptionFunc
	if input.Args.BaseURL != "" {
		opts = append(opts, gitlab.WithBaseURL(input.Args.BaseURL))
	}

	// TODO: add switch based on auth type
	git, err := gitlab.NewBasicAuthClient(
		input.Args.Auth.Basic.Username,
		input.Args.Auth.Basic.Password,
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "while creating GitLab API client")
	}

	req, err := git.NewRequest(input.Args.Method, input.Args.Path, input.Args.RequestBody, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while creating request")
	}

	req.WithContext(ctx)

	rawBody := map[string]interface{}{}
	_, err = git.Do(req, &rawBody)
	if err != nil {
		return nil, errors.Wrap(err, "while executing request")
	}

	artifact, err := r.renderOutput(input.Args.Output.GoTemplate, rawBody)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering additional data")
	}

	if err := r.saveOutput(artifact); err != nil {
		return nil, err
	}

	return &runner.WaitForCompletionOutput{
		Succeeded: true,
	}, nil
}

// Name returns the runner name.
func (r *RESTRunner) Name() string {
	return "gitlab.rest.api.v4"
}

// InjectLogger sets the logger on the runner.
func (r *RESTRunner) InjectLogger(logger *zap.Logger) {
	r.log = logger
}

func (r *RESTRunner) saveOutput(data []byte) error {
	if data == nil {
		return nil
	}

	r.log.Debug("Saving additional output", zap.String("path", r.cfg.Output.AdditionalFilePath))
	err := runner.SaveToFile(r.cfg.Output.AdditionalFilePath, data)
	if err != nil {
		return errors.Wrap(err, "while saving default output")
	}

	return nil
}

func (r *RESTRunner) readInputData(in runner.StartInput) (Input, error) {
	var args Arguments
	err := yaml.Unmarshal(in.Args, &args)
	if err != nil {
		return Input{}, errors.Wrap(err, "while unmarshalling runner arguments")
	}

	return Input{
		Args: args,
		Ctx:  in.RunnerCtx,
	}, nil
}

func (r *RESTRunner) renderOutput(artifactTemplate string, data map[string]interface{}) ([]byte, error) {
	if artifactTemplate == "" {
		return []byte{}, nil
	}

	tmpl, err := template.New("output").Parse(artifactTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load template")
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, data); err != nil {
		return nil, errors.Wrap(err, "while rendering output")
	}
	return buff.Bytes(), nil
}
