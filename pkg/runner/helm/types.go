package helm

import (
	"encoding/json"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"projectvoltron.dev/voltron/pkg/runner"
)

type outputter interface {
	ProduceHelmRelease(repository string, helmRelease *release.Release) ([]byte, error)
	ProduceAdditional(args Arguments, chrt *chart.Chart, rel *release.Release) ([]byte, error)
}

// Config holds Runner related configuration.
type Config struct {
	Command             CommandType
	HelmReleasePath     string `envconfig:"optional"`
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
	UpgradeCommandType = "upgrade"
)

type Arguments struct {
	CommonArgs
	InstallArgs
	UpgradeArgs
}

type CommonArgs struct {
	Values         map[string]interface{} `json:"values"`
	ValuesFromFile string                 `json:"valuesFromFile"`
	NoHooks        bool                   `json:"noHooks"`
	Chart          Chart                  `json:"chart"`
	Output         OutputArgs             `json:"output"`
}

type InstallArgs struct {
	Name         string `json:"name"`
	GenerateName bool   `json:"generateName"`
	Replace      bool   `json:"replace"`
}

type UpgradeArgs struct {
	ReuseValues bool `json:"reuseValues"`
	ResetValues bool `json:"resetValues"`
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
