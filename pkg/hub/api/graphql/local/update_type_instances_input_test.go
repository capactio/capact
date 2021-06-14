package graphql

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestUpdateTypeInstanceInputMarshalJSON(t *testing.T) {
	tests := map[string]struct {
		name      string
		rawInput  string
		expOutput string
	}{
		"Should handle empty typeInstance entry": {
			rawInput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				ownerID: null
				typeInstance: {}
				`),
			expOutput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				typeInstance: {}
				`),
		},
		"Should handle empty typeInstance.attributes entry": {
			rawInput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				ownerID: null
				typeInstance:
				  value:
				    parent: true
				`),
			expOutput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				typeInstance:
				  value:
				    parent: true
				`),
		},
		"Should handle null typeInstance.attributes entry": {
			rawInput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				ownerID: null
				typeInstance:
				  attributes: null
				  value:
				    parent: true
				`),
			expOutput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				typeInstance:
				  value:
				    parent: true
				`),
		},
		"Should handle populated typeInstance.attributes entry": {
			rawInput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				ownerID: null
				typeInstance:
				  attributes:
				    - path: cap.type.sample
				      revision: 0.1.1
				  value:
				    parent: true
				`),
			expOutput: heredoc.Doc(`
				id: 99e4e977-69fe-49a1-b0ae-f47f5a34153b
				typeInstance:
				  attributes:
				    - path: cap.type.sample
				      revision: 0.1.1
				  value:
				    parent: true
				`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			edited := UpdateTypeInstancesInput{}
			err := yaml.Unmarshal([]byte(tc.rawInput), &edited)
			require.NoError(t, err)

			// then
			out, err := yaml.Marshal(edited)
			require.NoError(t, err)

			assert.YAMLEq(t, tc.expOutput, string(out))
		})
	}
}
