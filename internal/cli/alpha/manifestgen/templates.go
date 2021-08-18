package manifestgen

import (
	_ "embed"

	"github.com/alecthomas/jsonschema"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

var (
	//go:embed templates/interface-group.yaml.tmpl
	interfaceGroupManifestTemplate string

	//go:embed templates/interface.yaml.tmpl
	interfaceManifestTemplate string

	//go:embed templates/type.yaml.tmpl
	typeManifestTemplate string

	//go:embed templates/output-type.yaml.tmpl
	outputTypeManifestTemplate string

	//go:embed templates/terraform-implementation.yaml.tmpl
	terraformImplementationManifestTemplate string

	//go:embed templates/helm-implementation.yaml.tmpl
	helmImplementationManifestTemplate string
)

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
	JSONSchema *jsonschema.Type
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

	ValuesYAML string
}
