package examples

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func Test(t *testing.T) {
	// Temp hack, as the $ref in schemas is a relative path.
	// In the future this test will be removed and replaced with
	// a ocftool validate command executed against all examples.
	// TODO: Remove as a part of https://cshark.atlassian.net/browse/SV-21
	mustChDirToRoot(t)

	tests := []struct {
		name       string
		jsonSchema string
		jsonObject string
	}{
		{
			name:       "Type Example should be valid",
			jsonSchema: "pkg/apis/0.0.1/schema/type.json",
			jsonObject: "pkg/apis/0.0.1/examples/type.json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", test.jsonSchema))
			documentLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", test.jsonObject))

			schema, err := gojsonschema.NewSchema(schemaLoader)
			require.NoError(t, err, "while creating schema validator")

			result, err := schema.Validate(documentLoader)
			require.NoError(t, err, "while validating object against JSON Schema")

			assertResultIsValid(t, result)
		})
	}
}

func assertResultIsValid(t *testing.T, result *gojsonschema.Result) {
	if !assert.True(t, result.Valid()) {
		t.Errorf("The document is not valid. see errors:\n")
		for _, desc := range result.Errors() {
			t.Errorf("- %s\n", desc.String())
		}
	}
}

func mustChDirToRoot(t *testing.T) {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../..")
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err.Error())
	}
}
