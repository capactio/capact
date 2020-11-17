// +build ocfmanifests

package examples

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

func TestManifestsValid(t *testing.T) {
<<<<<<< HEAD
<<<<<<< HEAD
	// Load the common schemas. Currently, the https $ref is not working as we didn't publish the spec yet.
	sl := gojsonschema.NewSchemaLoader()

	schemaRefPaths := []string{
		"../schema/common/json-schema-type.json",
		"../schema/common/type-ref.json",
		"../schema/common/input-type-instances.json",
		"../schema/common/output-type-instances.json",
		"../schema/common/metadata.json",
		"../schema/common/metadata-tags.json",
	}
	err := loadCommonSchemas(sl, schemaRefPaths)
	require.NoError(t, err, "while loading common schemas")
=======
	validator, err := ocftool.NewFilesystemManifestValidator("../..")
	require.NoError(t, err, "while creating validator instance")
>>>>>>> ef5e95c... working cli
=======
	validator := ocftool.NewFilesystemManifestValidator("../..")
>>>>>>> f4cdb86... fix tests

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
			fp, err := os.Open(tc.manifestPath)
			require.NoError(t, err, "while reading manifest test file")
			defer fp.Close()

			// when
			result, err := validator.ValidateYaml(fp)

			// then
			require.NoError(t, err, "while validating object against JSON Schema")
			require.True(t, result.Valid(), "is not valid")
		})
	}
}
