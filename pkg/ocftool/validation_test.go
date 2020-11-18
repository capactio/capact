package ocftool_test

import (
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
			manifestPath: "testdata/wrong_implementation.yaml",
			result:       false,
		},
		"Implementation should be valid": {
			manifestPath: "testdata/correct_implementation.yaml",
			result:       true,
		},
		"Should handle Go templates": {
			manifestPath: "testdata/correct_implementation_with_template.yaml",
			result:       true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given

			// when
			result := validator.ValidateFile(tc.manifestPath)

			// then
			require.Equal(t, tc.result, result.Valid(), result.Errors)
		})
	}
}
