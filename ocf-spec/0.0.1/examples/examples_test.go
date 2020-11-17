// +build ocfmanifests

package examples

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestManifestsValid in the future will be removed and replaced with
// an `ocftool validate` command executed against all examples.
// TODO: Remove as a part of https://cshark.atlassian.net/browse/SV-21
func TestManifestsValid(t *testing.T) {
	// Load the common schemas. Currently, the https $ref is not working as we didn't publish the spec yet.
	sl := gojsonschema.NewSchemaLoader()

	schemaRefPaths := []string{
		"../schema/common/json-schema-type.json",
		"../schema/common/metadata.json",
		"../schema/common/metadata-tags.json",
	}
	err := loadCommonSchemas(sl, schemaRefPaths)
	require.NoError(t, err, "while loading common schemas")

	tests := map[string]struct {
		jsonSchemaPath string
		manifestPath   string
	}{
		"Type Example should be valid": {
			jsonSchemaPath: "../schema/type.json",
			manifestPath:   "type.yaml",
		},
		"Tag Example should be valid": {
			jsonSchemaPath: "../schema/tag.json",
			manifestPath:   "tag.yaml",
		},
		"Vendor Example should be valid": {
			jsonSchemaPath: "../schema/vendor.json",
			manifestPath:   "vendor.yaml",
		},
		"RepoMetadata Example should be valid": {
			jsonSchemaPath: "../schema/repo-metadata.json",
			manifestPath:   "repo-metadata.yaml",
		},
		"InterfaceGroup Example should be valid": {
			jsonSchemaPath: "../schema/interface-group.json",
			manifestPath:   "interface-group.yaml",
		},
		"Interface Example should be valid": {
			jsonSchemaPath: "../schema/interface.json",
			manifestPath:   "interface.yaml",
		},
		"Implementation Example should be valid": {
			jsonSchemaPath: "../schema/implementation.json",
			manifestPath:   "implementation.yaml",
		},
		"TypeInstance Example should be valid": {
			jsonSchemaPath: "../schema/type-instance.json",
			manifestPath:   "type-instance.yaml",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", tc.jsonSchemaPath))
			schema, err := sl.Compile(schemaLoader)
			require.NoError(t, err, "while creating schema validator")

			manifest, err := documentLoader(tc.manifestPath)
			require.NoError(t, err, "while loading manifest")

			// when
			result, err := schema.Validate(manifest)
			require.NoError(t, err, "while validating object against JSON Schema")

			// then
			assertResultIsValid(t, result)
		})
	}
}

func documentLoader(path string) (gojsonschema.JSONLoader, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	obj := map[string]interface{}{}
	if err := yaml.Unmarshal(buf, &obj); err != nil {
		return nil, err
	}

	return gojsonschema.NewGoLoader(obj), nil
}

func assertResultIsValid(t *testing.T, result *gojsonschema.Result) {
	t.Helper()

	if !assert.True(t, result.Valid()) {
		t.Errorf("The document is not valid. see errors:\n")
		for _, desc := range result.Errors() {
			t.Errorf("- %s\n", desc.String())
		}
	}
}

func loadCommonSchemas(schemaLoader *gojsonschema.SchemaLoader, paths []string) error {
	for _, path := range paths {
		jsonLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", path))
		err := schemaLoader.AddSchemas(jsonLoader)
		if err != nil {
			return err
		}
	}

	return nil
}
