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
		// not sure, probably for InterfaceGroup the suffix has to be also modified, didn't check that far
	}

	return fmt.Sprintf("cap.%s.%s", strings.ToLower(string(manifestType)), suffix)
}

// AddRevisionToPath adds revision to manifest path.
func AddRevisionToPath(path string, revision string) string {
	return path + ":" + revision
}

// GetDefaultMetadata creates a new Metadata object and sets default values.
func GetDefaultMetadata() types.ImplementationMetadata {
	var metadata types.ImplementationMetadata
	metadata.DocumentationURL = ptr.String("https://example.com")
	metadata.SupportURL = ptr.String("https://example.com")
	metadata.IconURL = ptr.String("https://example.com/icon.png")
	metadata.Maintainers = []types.Maintainer{
		{
			Email: "dev@example.com",
			Name:  ptr.String("Example Dev"),
			URL:   ptr.String("https://example.com"),
		},
	}
	metadata.License.Name = ApacheLicense
	return metadata
}
