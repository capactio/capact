package helm

import "projectvoltron.dev/voltron/pkg/runner"

// Config holds Runner related configuration.
type Config struct {
	HelmDriver          string `envconfig:"default=secrets"`
	RepositoryCachePath string `envconfig:"default=/tmp/helm"`
}

type CommandType string

const (
	InstallCommandType = "install"
)

type Arguments struct {
	Command        CommandType            `json:"command"`
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
	Directory  string           `json:"directory"`
	Default    DefaultOutput    `json:"default"`
	Additional AdditionalOutput `json:"additional"`
}

type DefaultOutput struct {
	FileName string `json:"fileName"`
}

type AdditionalOutput struct {
	FileName string `json:"fileName"`
	Value    string `json:"value"`
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
	Args    Arguments
	ExecCtx runner.ExecutionContext
}

type Status struct {
	Succeeded bool
	Message   string
}

type Output struct {
	Default    File
	Additional *File
}

type File struct {
	Path  string
	Value []byte
}
