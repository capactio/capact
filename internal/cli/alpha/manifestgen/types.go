package manifestgen

import (
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// Config stores generic input parameters for content generation
type Config struct {
	ManifestPath     string
	ManifestRevision string
}

// InterfaceConfig stores input parameters for Interface content generation
type InterfaceConfig struct {
	Config
}

// ImplementationConfig stores input parameters for Implementation content generation
type ImplementationConfig struct {
	Config
	InterfacePathWithRevision string
}

// TerraformConfig stores input parameters for Terraform-based Implementation content generation
type TerraformConfig struct {
	ImplementationConfig

	ModulePath      string
	ModuleSourceURL string
	Provider        Provider
}

// HelmConfig stores input parameters for Helm-based Implementation content generation.
type HelmConfig struct {
	ImplementationConfig

	ChartName    string
	ChartRepoURL string
	ChartVersion string
}

type templatingConfig struct {
	Template string
	Input    interface{}
}

type templatingInput struct {
	Name     string
	Prefix   string
	Revision string
}

type interfaceGroupTemplatingInput struct {
	templatingInput
}

type interfaceTemplatingInput struct {
	templatingInput
}

type outputTypeTemplatingInput struct {
	templatingInput
}

type typeTemplatingInput struct {
	templatingInput
	JSONSchema string
}

type terraformImplementationTemplatingInput struct {
	templatingInput

	InterfacePath     string
	InterfaceRevision string
	ModuleSourceURL   string
	Outputs           []*tfconfig.Output
	Provider          Provider
	Variables         []*tfconfig.Variable
}

type helmImplementationTemplatingInput struct {
	templatingInput

	InterfacePath     string
	InterfaceRevision string

	HelmChartName    string
	HelmChartVersion string
	HelmRepoURL      string

	ValuesYAML  string
	ArgsWarning string
}
