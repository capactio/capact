package glabapi

import "capact.io/capact/pkg/runner"

// Config holds RESTRunner related configuration.
type Config struct {
	Output struct {
		// Extracting resource data from response body
		AdditionalFilePath string `envconfig:"default=/tmp/additional.yaml"`
	}
}

// Input stores the input configuration for the runner.
type Input struct {
	Args Arguments
	Ctx  runner.Context
}

// Arguments stores the input arguments for the GitLab API runner operation.
type Arguments struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	RequestBody map[string]interface{} `json:"body"`
	BaseURL     string                 `json:"baseURL"`
	Auth        Auth                   `json:"auth"`
	Output      OutputArgs             `json:"output"`
}

// Auth holds auth data for GitLab API
// TODO: add token support
type Auth struct {
	Basic struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"basic"`
}

// OutputArgs stores input arguments for generating the output artifacts.
type OutputArgs struct {
	GoTemplate string `json:"goTemplate"`
}
