package common

import (
	"capact.io/capact/internal/cli/alpha/manifestgen"
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
	ApacheLicense = "Apache 2.0"
	// AWSProvider hold a name for AWS Provider
	AWSProvider = "AWS"
	// AttributeManifest hold a name for Attribute Manifest
	AttributeManifest = "Attribute"
	// GCPProvider hold a name for GCP Provider
	GCPProvider = "GCP"
	// InterfaceManifest hold a name for Interface Manifest
	InterfaceManifest = "Interface"
	// InterfaceGroupManifest hold a name for InterfaceGroup Manifest
	InterfaceGroupManifest = "InterfaceGroup"
	// ImplementationManifest hold a name for Implementation Manifest
	ImplementationManifest = "Implementation"
	// TypeManifest hold a name for Type Manifest
	TypeManifest = "Type"
)
