package common

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// Metadata gathers common Metadata information for all manifest types.
type Metadata = types.ImplementationMetadata

// ManifestGenOptions is a struct based on which manifests are generated.
type ManifestGenOptions struct {
	Directory      string
	InterfacePath  string
	ManifestsKinds []string
	ManifestPath   string
	Metadata       Metadata
	Overwrite      bool
	Revision       string
	TypeInputPath  types.ManifestRef
	TypeOutputPath types.ManifestRef
}

var (
	// ApacheLicense hold a name for Apache License.
	ApacheLicense = ptr.String("Apache 2.0")
)
