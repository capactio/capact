package manifestgen

import (
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// Config stores generic input parameters for content generation
type Config struct {
	ManifestPath     string
	ManifestRevision string
	ManifestMetadata MetaDataInfo
}

// AttributeConfig stores input parameters for Attribute content generation
type AttributeConfig struct {
	Config
}

// InterfaceConfig stores input parameters for Interface content generation
type InterfaceConfig struct {
	Config
	InputPathWithRevision  string
	OutputPathWithRevision string
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

// EmptyImplementationConfig stores input parameters for empty Implementation content generation.
type EmptyImplementationConfig struct {
	ImplementationConfig
}

type templatingConfig struct {
	Template string
	Input    interface{}
}

type templatingInput struct {
	Name     string
	Prefix   string
	Revision string
	Metadata MetaDataInfo
}

type attributeTemplatingInput struct {
	templatingInput
}

type interfaceGroupTemplatingInput struct {
	templatingInput
}

type interfaceTemplatingInput struct {
	templatingInput
	InputTypeName      string
	InputTypeRevision  string
	OutputTypeName     string
	OutputTypeRevision string
}

type outputTypeTemplatingInput struct {
	templatingInput
}

type typeTemplatingInput struct {
	templatingInput
	JSONSchema string
}

type emptyImplementationTemplatingInput struct {
	templatingInput
	InterfacePath     string
	InterfaceRevision string
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
