package common

// CreateManifestPath create a manifest path based on a manifest type and suffix.
func CreateManifestPath(manifestType string, suffix string) string {
	suffixes := map[string]string{
		AttributeManifest:      "attribute",
		TypeManifest:           "type",
		InterfaceManifest:      "interface",
		InterfaceGroupManifest: "interfaceGroup",
		ImplementationManifest: "implementation",
	}
	return "cap." + suffixes[manifestType] + "." + suffix
}

// AddRevisionToPath adds revision to manifest path.
func AddRevisionToPath(path string, revision string) string {
	return path + ":" + revision
}

// GetDefaultMetadata creates a new Metadata object and sets default values.
func GetDefaultMetadata() Metadata {
	var metadata Metadata
	metadata.DocumentationURL = "https://example.com"
	metadata.SupportURL = "https://example.com"
	metadata.IconURL = "https://example.com/icon.png"
	metadata.Maintainers = []Maintainers{
		{
			Email: "dev@example.com",
			Name:  "Example Dev",
			URL:   "https://example.com",
		},
	}
	metadata.License.Name = &ApacheLicense
	return metadata
}
