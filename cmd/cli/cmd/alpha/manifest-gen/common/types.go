package common

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// ManifestGenOptions is a struct based on which manifests are generated.
type ManifestGenOptions struct {
	Directory      string
	InterfacePath  string
	ManifestsType  []string
	ManifestPath   string
	Metadata       types.ImplementationMetadata
	Overwrite      bool
	Revision       string
	TypeInputPath  string
	TypeOutputPath string
}

var (
	// ApacheLicense hold a name for Apache License.
	ApacheLicense = ptr.String("Apache 2.0")
)
