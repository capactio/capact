package graphql

import (
	"fmt"
	"strings"
)

// ManifestReference holds Capact OCF manifest identification details.
type ManifestReference struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

// GQLQueryName returns name of GraphQL query needed to get details of the manifest.
// TODO: Very naive implementation. To refactor for later once it is more widely used.
func (r ManifestReference) GQLQueryName() (string, error) {
	parts := strings.Split(r.Path, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("path parts for %q cannot be less than 3", r.Path)
	}

	if parts[1] == "core" {
		return parts[2], nil
	}

	return parts[1], nil
}
