package examples

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestExampleSuccess in the future will be removed and replaced with
// a ocftool validate command executed against all examples.
// TODO: Remove as a part of https://cshark.atlassian.net/browse/SV-21
func TestExampleSuccess(t *testing.T) {
	// Temp hack, as the $ref in schemas is a relative path.
	// We need to wait with 'https' $ref until the schema will be public.
	mustChDirToRootVersion(t)

	tests := map[string]struct {
		jsonSchemaPath string
		manifestPath   string
	}{
		"Type Example should be valid": {
			jsonSchemaPath: "schema/type.json",
			manifestPath:   "examples/type.yaml",
		},
		"Tag Example should be valid": {
			jsonSchemaPath: "schema/tag.json",
			manifestPath:   "examples/tag.yaml",
		},
		"Vendor Example should be valid": {
			jsonSchemaPath: "schema/vendor.json",
			manifestPath:   "examples/vendor.yaml",
		},
		"RepoMetadata Example should be valid": {
			jsonSchemaPath: "schema/repo-metadata.json",
			manifestPath:   "examples/repo-metadata.yaml",
		},
		"Interface Example should be valid": {
			jsonSchemaPath: "schema/interface.json",
			manifestPath:   "examples/interface.yaml",
		},
		"Implementation Example should be valid": {
			jsonSchemaPath: "schema/implementation.json",
			manifestPath:   "examples/implementation.yaml",
		},
		"TypeInstance Example should be valid": {
			jsonSchemaPath: "schema/type-instance.json",
			manifestPath:   "examples/type-instance.yaml",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", tc.jsonSchemaPath))
			schema, err := gojsonschema.NewSchema(schemaLoader)
			require.NoError(t, err, "while creating schema validator")

			manifest, err := documentLoader(tc.manifestPath)
			require.NoError(t, err, "while loading manifest")

			// set root schema name?
			result, err := schema.Validate(manifest)
			require.NoError(t, err, "while validating object against JSON Schema")

			assertResultIsValid(t, result)
		})
	}
}

// The NewBytesLoader is used.
// Other option is to unmarshal to map[string]interface{} and use the NewGoLoader
// but we need to deal with the diff between JSON and YAML manually.
// For now is enough to use a external lib for doing that.
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

func mustChDirToRootVersion(t *testing.T) {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err.Error())
	}
}
