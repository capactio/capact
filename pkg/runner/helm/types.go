package helm

import (
	"encoding/json"

	"projectvoltron.dev/voltron/pkg/runner"
)

// Config holds Runner related configuration.
type Config struct {
	Command             CommandType
	HelmDriver          string `envconfig:"default=secrets"`
	RepositoryCachePath string `envconfig:"default=/tmp/helm"`
	Output              struct {
		HelmReleaseFilePath string `envconfig:"default=/tmp/helm-release.yaml"`
		// Extracting resource metadata from Kubernetes as outputs
		AdditionalFilePath string `envconfig:"default=/tmp/additional.yaml"`
	}
}

type CommandType string

const (
	InstallCommandType = "install"
)

type Arguments struct {
	Name           string                 `json:"name"`
	Chart          Chart                  `json:"chart"`
	Values         map[string]interface{} `json:"values"`
	ValuesFromFile string                 `json:"valuesFromFile"`
	NoHooks        bool                   `json:"noHooks"`
	Replace        bool                   `json:"replace"`
	GenerateName   bool                   `json:"generateName"`

	Output OutputArgs `json:"output"`
}

type OutputArgs struct {
	GoTemplate json.RawMessage `json:"goTemplate"`
}

type Chart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
}

type ChartRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     Chart  `json:"chart"`
}

type Input struct {
	Args Arguments
	Ctx  runner.Context
}

type Status struct {
	Succeeded bool
	Message   string
}

type Output struct {
	Release    []byte
	Additional []byte
}
