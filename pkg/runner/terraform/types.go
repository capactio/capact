package terraform

import "encoding/json"

// CommandType represents the operation type to be performed by the runner.
type CommandType string

const (
	// ApplyCommand is an operation to perform terraform apply
	ApplyCommand = "apply"
	// DestroyCommand is an operation to perform terraform destroy
	DestroyCommand = "destroy"
	// PlanCommand is an operation to perform terraform plan
	PlanCommand = "plan"
)

// Arguments stores the input arguments for the runner operation.
type Arguments struct {
	Command   string           `yaml:"command"`
	Name      string           `yaml:"name"`
	Module    Module           `yaml:"module"`
	Env       []string         `yaml:"env"`
	Variables string           `yaml:"variables"`
	Output    AdditionalOutput `yaml:"output"`
	// TODO destroy needs tfstate file
}

// Module stores the source details of the Terraform module.
type Module struct {
	Name   string
	Source string
}

// AdditionalOutput stores input arguments for generating the additional output.
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

// OutputConfig stores the configuration for the generated output file.
type OutputConfig struct {
	TerraformReleaseFilePath string `envconfig:"default=/tmp/terraform-release.yaml"`
	AdditionalFilePath       string `envconfig:"default=/tmp/additional.yaml"`
	TfstateFilePath          string `envconfig:"default=/tmp/terraform.tfstate"`
}

// Release stores the details about the Terraform release.
type Release struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// StateTypeInstance stores the details about the Terraform state TypeInstance
type StateTypeInstance struct {
	State     []byte `json:"state"`
	Variables []byte `json:"variables"`
}

// Output stores the generated output artifacts.
type Output struct {
	Release    []byte
	Additional []byte
	State      StateTypeInstance
}
