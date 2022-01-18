package manifestgen

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// ManifestPath is a type for manifest path.
type ManifestPath string

// ManifestContent is a type for manifest content.
type ManifestContent []byte

// ManifestCollection is a type for manifest collections.
type ManifestCollection map[ManifestPath]ManifestContent

// Config stores generic input parameters for content generation.
type Config struct {
	ManifestRef types.ManifestRef
}

// AttributeConfig stores input parameters for Attribute content generation.
type AttributeConfig struct {
	Config
	Metadata types.InterfaceMetadata
}

// InterfaceConfig stores input parameters for Interface content generation.
type InterfaceConfig struct {
	Config
	Metadata      types.InterfaceMetadata
	InputTypeRef  string
	OutputTypeRef string
}

// InterfaceGroupConfig stores input parameters for InterfaceGroup content generation.
type InterfaceGroupConfig struct {
	Config
	Metadata types.InterfaceMetadata
}

// ImplementationConfig stores input parameters for Implementation content generation.
type ImplementationConfig struct {
	Config
	Metadata                  types.ImplementationMetadata
	InterfacePathWithRevision string
}

// TerraformConfig stores input parameters for Terraform-based Implementation content generation.
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
	AdditionalInputTypeName string
	ImplementationConfig
	GenerateInputType bool
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

type attributeTemplatingInput struct {
	templatingInput
	Metadata types.InterfaceMetadata
}

type interfaceGroupTemplatingInput struct {
	templatingInput
	Metadata types.InterfaceMetadata
}

type interfaceTemplatingInput struct {
	templatingInput
	Metadata  types.InterfaceMetadata
	InputRef  types.ManifestRef
	OutputRef types.ManifestRef
}

type typeTemplatingInput struct {
	templatingInput
	Metadata   types.TypeMetadata
	JSONSchema string
}

type emptyImplementationTemplatingInput struct {
	templatingInput
	Metadata            types.ImplementationMetadata
	AdditionalInputName string
	InterfaceRef        types.ManifestRef
}

type terraformImplementationTemplatingInput struct {
	templatingInput

	InterfaceRef    types.ManifestRef
	Metadata        types.ImplementationMetadata
	ModuleSourceURL string
	Outputs         []*tfconfig.Output
	Provider        Provider
	Variables       []*tfconfig.Variable
}

type helmImplementationTemplatingInput struct {
	templatingInput

	Metadata types.ImplementationMetadata

	InterfaceRef types.ManifestRef

	HelmChartName    string
	HelmChartVersion string
	HelmRepoURL      string

	ValuesYAML  string
	ArgsWarning string
}
