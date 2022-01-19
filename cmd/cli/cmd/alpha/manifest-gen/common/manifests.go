package common

import (
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// CreateManifestPath create a manifest path based on a manifest type and suffix.
func CreateManifestPath(manifestType types.ManifestKind, suffix string) string {
	if manifestType == types.InterfaceGroupManifestKind {
		// InterfaceGroup resides in the same directory as Interfaces
		manifestType = types.InterfaceManifestKind
	}

	return fmt.Sprintf("cap.%s.%s", strings.ToLower(string(manifestType)), suffix)
}

// AddRevisionToPath adds revision to manifest path.
func AddRevisionToPath(path string, revision string) string {
	return path + ":" + revision
}

// GetDefaultInterfaceMetadata creates a new Metadata object for Interface-kind manifests and sets default values.
func GetDefaultInterfaceMetadata() types.InterfaceMetadata {
	return types.InterfaceMetadata{
		DocumentationURL: defaultURL(),
		SupportURL:       defaultURL(),
		IconURL:          defaultIconURL(),
		Maintainers:      defaultMaintainers(),
	}
}

// GetDefaultImplementationMetadata creates a new Metadata object for Implementation and sets default values.
func GetDefaultImplementationMetadata() types.ImplementationMetadata {
	return types.ImplementationMetadata{
		DocumentationURL: defaultURL(),
		SupportURL:       defaultURL(),
		IconURL:          defaultIconURL(),
		Maintainers:      defaultMaintainers(),
		License:          defaultLicense(),
	}
}

func defaultURL() *string {
	return ptr.String("https://example.com")
}

func defaultIconURL() *string {
	return ptr.String("https://example.com/icon.png")
}

func defaultMaintainers() []types.Maintainer {
	return []types.Maintainer{
		{
			Email: "dev@example.com",
			Name:  ptr.String("Example Dev"),
			URL:   ptr.String("https://example.com"),
		},
	}
}

func defaultLicense() types.License {
	return types.License{
		Name: ApacheLicense,
	}
}
