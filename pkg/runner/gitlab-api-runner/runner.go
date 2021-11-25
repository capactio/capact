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

	git, err := r.gitlabClientForAuth(input.Args.Auth, opts)
	if err != nil {
		return nil, errors.Wrap(err, "while creating GitLab API client")
	}

	req, err := git.NewRequest(input.Args.Method, input.Args.Path, input.Args.RequestBody, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while creating request")
	}

	req.WithContext(ctx)

	if input.Args.QueryParameters != nil {
		req.URL.RawQuery = input.Args.QueryParameters.Encode()
	}

	var rawBody interface{}
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

func (r *RESTRunner) gitlabClientForAuth(auth Auth, opts []gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	token := auth.Token
	basic := auth.Basic

	if token != nil && basic != nil {
		return nil, errors.New("both token and basic credentials must not be provided")
	}

	// access token
	if token != nil {
		return gitlab.NewClient(
			*token,
			opts...,
		)
	}

	// basic credentials
	if basic != nil {
		return gitlab.NewBasicAuthClient(
			basic.Username,
			basic.Password,
			opts...,
		)
	}

	return nil, errors.New("no token or basic credentials provided")
}

func (r *RESTRunner) renderOutput(artifactTemplate string, data interface{}) ([]byte, error) {
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
