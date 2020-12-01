package manifest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/validate"
	"projectvoltron.dev/voltron/pkg/sdk/manifest"
)

func TestFilesystemValidator_ValidateFile(t *testing.T) {
	validator := manifest.NewFilesystemValidator(&validate.LocalSchema{}, "../../../ocf-spec")

	tests := map[string]struct {
		manifestPath string
		result       bool
	}{
		"Implementation should be invalid": {
			manifestPath: "testdata/wrong_implementation.yaml",
			result:       false,
		},
		"Implementation should be valid": {
			manifestPath: "testdata/correct_implementation.yaml",
			result:       true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given

			// when
			result, err := validator.ValidateFile(tc.manifestPath)

			// then
			require.Nil(t, err, "failed to read file: %v", err)
			require.Equal(t, tc.result, result.Valid(), result.Errors)
		})
	}
}
