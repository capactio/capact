package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimSSHKey(t *testing.T) {
	//given
	tests := map[string]struct {
		givenPath    string
		expectedPath string
	}{
		"Trim SSH key query string": {
			givenPath:    "git@github.com:example/hub-manifests.git?sshkey=LS0t...S0K",
			expectedPath: "git@github.com:example/hub-manifests.git",
		},
		"Trim SSH key query string with additional parameters": {
			givenPath:    "git@github.com:example/hub-manifests.git?sshkey=LS0t...S0K&ref=ref_branch",
			expectedPath: "git@github.com:example/hub-manifests.git&ref=ref_branch",
		},
		"Trim SSH key query passed as ampersand operator": {
			givenPath:    "git@github.com:example/hub-manifests.git?ref=ref_branch&sshkey=LS0t...S0K",
			expectedPath: "git@github.com:example/hub-manifests.git?ref=ref_branch",
		},
		"Trim SSH key without SSH key": {
			givenPath:    "git@github.com:example/hub-manifests.git?ref=ref_branch",
			expectedPath: "git@github.com:example/hub-manifests.git?ref=ref_branch",
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			trimmedSSHKey := trimSSHKey(tc.givenPath)

			// then
			assert.Equal(t, tc.expectedPath, trimmedSSHKey)
		})
	}
}
