package controller

import (
	"gotest.tools/assert"
	"testing"
	"time"
)

func Test(t *testing.T) {
	tests := map[string]struct {
		givenPath string
		expPath   string
	}{
		"Should add `-local1` suffix": {
			givenPath: "cap.interface.productivity.jira.install",
			expPath:   "cap.interface.productivity.jira.install-local",
		},
		"Should not add additional `-local` suffix": {
			givenPath: "cap.interface.productivity.jira.install-local",
			expPath:   "cap.interface.productivity.jira.install-local",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			svc := NewActionService(nil, nil, "", time.Millisecond)

			// when
			gotPath := svc.ensureLocalSuffix(tc.givenPath)

			// then
			assert.Equal(t, tc.expPath, gotPath)
		})
	}
}
