package capact

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCRDLocationFromVersion(t *testing.T) {
	tests := map[string]struct {
		givenParams    *InputParameters
		expCRDLocation string
	}{
		"local version": {
			givenParams:    &InputParameters{Version: "@local"},
			expCRDLocation: LocalCRDPath,
		},
		"stable version": {
			givenParams:    &InputParameters{Version: "0.5.0"},
			expCRDLocation: fmt.Sprintf(CRDUrlFormat, "0.5.0"),
		},
		"latest version": {
			givenParams:    &InputParameters{Version: "0.5.0-67e2484"},
			expCRDLocation: fmt.Sprintf(CRDUrlFormat, "67e2484"),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			err := tc.givenParams.resolveCRDLocationFromVersion()

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expCRDLocation, tc.givenParams.ActionCRDLocation)
		})
	}
}
