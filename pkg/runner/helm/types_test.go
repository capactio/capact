package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestDefaultArguments(t *testing.T) {
	tests := map[string]struct {
		givenArgs     []byte
		expMaxHistory int
	}{
		"Should set default to MaxHistory": {
			givenArgs:     []byte(`{}`),
			expMaxHistory: MaxHistoryDefault,
		},
		"Should override MaxHistory": {
			givenArgs:     []byte(`{"maxHistory": 12}`),
			expMaxHistory: 12,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			args := DefaultArguments()

			// when
			err := yaml.Unmarshal(tc.givenArgs, &args)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expMaxHistory, args.MaxHistory)
		})
	}
}
