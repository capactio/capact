package common

import (
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// Metadata is a alias for MetaDataInfo struct
type Metadata = manifestgen.MetaDataInfo

// Maintainers is a alias for Maintainer struct
type Maintainers = manifestgen.Maintainer

// ManifestGenOptions is a struct based on which manifests are generated
type ManifestGenOptions struct {
	Directory      string
	InterfacePath  string
	ManifestsType  []string
	ManifestPath   string
	Metadata       Metadata
	Overwrite      bool
	Revision       string
	TypeInputPath  string
	TypeOutputPath string
}

var (
	// ApacheLicense hold a name for Apache License
	ApacheLicense string = "Apache 2.0"
)

const (
	// AttributeManifest hold a name for Attribute Manifest
	AttributeManifest = string(types.AttributeManifestKind)
	// InterfaceManifest hold a name for Interface Manifest
	InterfaceManifest = string(types.InterfaceManifestKind)
	// InterfaceGroupManifest hold a name for InterfaceGroup Manifest
	InterfaceGroupManifest = string(types.InterfaceGroupManifestKind)
	// ImplementationManifest hold a name for Implementation Manifest
	ImplementationManifest = string(types.ImplementationManifestKind)
	// TypeManifest hold a name for Type Manifest
	TypeManifest = string(types.TypeManifestKind)
)
