package manifestgen

import (
	_ "embed"
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
