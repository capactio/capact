package helm

import (
	"capact.io/capact/pkg/runner"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type outputter interface {
	ProduceHelmRelease(repository string, helmRelease *release.Release) ([]byte, error)
	ProduceAdditional(args OutputArgs, chrt *chart.Chart, rel *release.Release) ([]byte, error)
}

// Config holds Runner related configuration.
type Config struct {
	OptionalKubeconfigTI string `envconfig:"optional"`
	Command              CommandType
	HelmReleasePath      string `envconfig:"optional"`
	HelmDriver           string `envconfig:"default=secrets"`
	RepositoryCachePath  string `envconfig:"default=/tmp/helm"`
	Output               struct {
		HelmReleaseFilePath string `envconfig:"default=/tmp/helm-release.yaml"`
		// Extracting resource metadata from Kubernetes as outputs
		AdditionalFilePath string `envconfig:"default=/tmp/additional.yaml"`
	}
}

// CommandType represents the operation type to be performed by the runner.
type CommandType string

const (
	// InstallCommandType is an operation to install a new Helm chart.
	InstallCommandType = "install"
	// UpgradeCommandType is an operation to upgrade an Helm release.
	UpgradeCommandType = "upgrade"
	// MaxHistoryDefault limits the maximum number of revisions saved per release.
	// Same value as defined by `helm upgrade` cmd: https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/cli/environment.go#L37-L38
	MaxHistoryDefault = 10
)

// DefaultArguments returns Helm Arguments with default values.
func DefaultArguments() Arguments {
	return Arguments{
		UpgradeArgs: UpgradeArgs{
			MaxHistory: MaxHistoryDefault,
		},
	}
}

// Arguments stores the input arguments for the Helm runner operation.
type Arguments struct {
	CommonArgs
	InstallArgs
	UpgradeArgs
}

// CommonArgs stores common arguments used in every operation.
type CommonArgs struct {
	Values         map[string]interface{} `json:"values"`
	ValuesFromFile string                 `json:"valuesFromFile"`
	NoHooks        bool                   `json:"noHooks"`
	Chart          Chart                  `json:"chart"`
	Output         OutputArgs             `json:"output"`
}

// InstallArgs stores input arguments to the install operation.
type InstallArgs struct {
	Name         string `json:"name"`
	GenerateName bool   `json:"generateName"`
	Replace      bool   `json:"replace"`
}

// UpgradeArgs stores input arguments for the upgrade operation.
type UpgradeArgs struct {
	ReuseValues bool `json:"reuseValues"`
	ResetValues bool `json:"resetValues"`
	MaxHistory  int  `json:"maxHistory"`
}

// OutputArgs stores input arguments for generating the output artifacts.
type OutputArgs struct {
	GoTemplate string `json:"goTemplate"`
}

// Chart represents a Helm chart.
type Chart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
}

// ChartRelease represents a Helm chart release.
type ChartRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     Chart  `json:"chart"`
}

// Input stores the input configuration for the runner.
type Input struct {
	Args Arguments
	Ctx  runner.Context
}

// Status indicates the status of the runner operation.
type Status struct {
	Succeeded bool
	Message   string
}

// Output stores the produces output artifacts.
type Output struct {
	Release    []byte
	Additional []byte
}
