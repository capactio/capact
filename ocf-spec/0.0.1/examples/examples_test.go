// +build ocfmanifests

package examples

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/schema"
	"capact.io/capact/pkg/sdk/validation/manifest"

	"github.com/stretchr/testify/require"
)

func TestManifestsValid(t *testing.T) {
	fs, ocfSchemaRootPath := schema.NewProvider("../..").FileSystem()
	validator := manifest.NewDefaultFilesystemValidator(fs, ocfSchemaRootPath)

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
		"Attribute should be valid": {
			manifestPath: "attribute.yaml",
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
			// when
			result, err := validator.Do(context.Background(), tc.manifestPath)

			// then
			require.Nil(t, err, "returned error: %v", err)
			require.True(t, result.Valid(), "is not valid, errors: %v", result.Errors)
		})
	}
}
