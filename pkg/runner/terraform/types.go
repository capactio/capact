package terraform

import "encoding/json"

type CommandType string

type Arguments struct {
	Command   string           `yaml:"command"`
	Name      string           `yaml:"name"`
	Module    Module           `yaml:"module"`
	Env       []string         `yaml:"env"`
	Variables string           `yaml:"variables"`
	Output    AdditionalOutput `yaml:"output"`
	// TODO destroy needs tfstate file
}

type Module struct {
	Name   string
	Source string
}

type AdditionalOutput struct {
	GoTemplate json.RawMessage `yaml:"goTemplate"`
}

// Config holds Runner related configuration.
type Config struct {
	WorkDir                   string `envconfig:"default=/workspace"`
	TerraformPath             string `envconfig:"default=terraform"`
	StateTypeInstanceFilepath string `envconfig:"optional"`
	Output                    OutputConfig
}

type OutputConfig struct {
	TerraformReleaseFilePath string `envconfig:"default=/tmp/terraform-release.yaml"`
	AdditionalFilePath       string `envconfig:"default=/tmp/additional.yaml"`
	TfstateFilePath          string `envconfig:"default=/tmp/terraform.tfstate"`
}

type Release struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type StateTypeInstance struct {
	State     []byte
	Variables []byte
}

type Output struct {
	Release    []byte
	Additional []byte
	State      StateTypeInstance
}
