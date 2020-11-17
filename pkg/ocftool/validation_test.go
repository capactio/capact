package ocftool_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

func TestVerifyValidator(t *testing.T) {
	validator := ocftool.NewFilesystemManifestValidator("../../ocf-spec")

	tests := map[string]struct {
		manifestPath string
		result       bool
	}{
		"Implementation should be invalid": {
			manifestPath: "test_manifests/wrong_implementation.yaml",
			result:       false,
		},
		"Implementation should be valid": {
			manifestPath: "test_manifests/correct_implementation.yaml",
			result:       true,
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
			require.Equal(t, tc.result, result.Valid(), result.Errors())
		})
	}
}
