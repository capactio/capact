// +build ocfmanifests

package examples

import (
	"testing"

	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

func TestManifestsValid(t *testing.T) {
	validator := ocftool.NewFilesystemManifestValidator("../..")

	tests := map[string]struct {
		manifestPath string
	}{
		"Implementation should be valid": {
			manifestPath: "implementation.yaml",
		},
		"InterfaceGroup should be valid": {
			manifestPath: "interface-group.yaml",
		},
		"Interface should be valid": {
			manifestPath: "interface.yaml",
		},
		"RepoMetadata should be valid": {
			manifestPath: "repo-metadata.yaml",
		},
		"Tag should be valid": {
			manifestPath: "tag.yaml",
		},
		"TypeInstance should be valid": {
			manifestPath: "type-instance.yaml",
		},
		"Type should be valid": {
			manifestPath: "type.yaml",
		},
		"Vendor should be valid": {
			manifestPath: "vendor.yaml",
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given

			// when
			result := validator.ValidateFile(tc.manifestPath)

			// then
			require.True(t, result.Valid(), "is not valid, errors: %v", result.Errors)
		})
	}
}
