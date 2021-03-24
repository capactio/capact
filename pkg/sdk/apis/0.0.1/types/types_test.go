package types_test

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type marshaler interface {
	Marshal() ([]byte, error)
}

func TestUnmarshalAndMarshalActionProduceSameResults(t *testing.T) {
	mustChDirToRoot(t)

	tests := map[string]struct {
		examplePath     string
		unmarshalMethod func(data []byte) (marshaler, error)
	}{
		"Implementation": {
			examplePath: "implementation.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalImplementation(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
		"Interface": {
			examplePath: "interface.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalInterface(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
		"RepoMetadata": {
			examplePath: "repo-metadata.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalRepoMetadata(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
		"Attribute": {
			examplePath: "attribute.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalAttribute(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
		"Type": {
			examplePath: "type.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalType(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
		"Vendor": {
			examplePath: "vendor.yaml",
			unmarshalMethod: func(data []byte) (marshaler, error) {
				obj, err := types.UnmarshalVendor(data)
				if err != nil {
					return nil, err
				}
				return &obj, nil
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			buf, err := ioutil.ReadFile(path.Join("./ocf-spec/0.0.1/examples/", tc.examplePath))
			require.NoError(t, err, "while reading example file")

			buf, err = yaml.YAMLToJSON(buf)
			require.NoError(t, err, "while converting YAML to JSON")

			// when
			gotUnmarshal, err := tc.unmarshalMethod(buf)
			require.NoError(t, err, "while unmarshaling example file")

			gotMarshal, err := gotUnmarshal.Marshal()
			require.NoError(t, err, "while marshaling example file")

			// then
			// TODO: we can have a missing field with that assertion, should be fixed later.
			assert.JSONEq(t, string(buf), string(gotMarshal))
		})
	}
}

func mustChDirToRoot(t *testing.T) {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../../")
	err := os.Chdir(dir)
	if err != nil {
		t.Fatal(err.Error())
	}
}
