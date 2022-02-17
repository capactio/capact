package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimLastNodeFromOCFPath(t *testing.T) {
	tests := map[string]struct {
		givenPath string
		expPath   string
	}{
		"Should remove last node name": {
			givenPath: "cap.core.type.examples.name",
			expPath:   "cap.core.type.examples",
		},
		"Should trim trailing separator": {
			givenPath: "cap.",
			expPath:   "cap",
		},
		"Should return given path if no separator detected": {
			givenPath: "cap",
			expPath:   "cap",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			gotPath := TrimLastNodeFromOCFPath(tc.givenPath)
			assert.Equal(t, tc.expPath, gotPath)
		})
	}
}
