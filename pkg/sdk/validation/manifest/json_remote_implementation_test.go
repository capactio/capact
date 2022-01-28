package manifest_test

import (
	"context"
	"errors"
	"testing"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/validation/manifest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckParentNodesAssociation(t *testing.T) {
	tests := map[string]struct {
		knownTypesByPathPattern map[string][]*gqlpublicapi.Type
		relationsToParentNode   manifest.ParentNodesAssociation
		expErrors               []error
	}{
		"should success as all nodes are attached to parent nodes": {
			knownTypesByPathPattern: map[string][]*gqlpublicapi.Type{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry"},
					{Path: "cap.type.platform.nomad"},
				},
			},

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					"cap.type.platform.cloud-foundry", "cap.type.platform.nomad",
				},
			},

			expErrors: nil, // no errors
		},
		"should detect that one Type is not attached to parent node (singular)": {
			knownTypesByPathPattern: map[string][]*gqlpublicapi.Type{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.nomad"},
				},
			},
			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					"cap.type.platform.cloud-foundry", "cap.type.platform.nomad",
				},
			},

			expErrors: []error{errors.New("Type cap.type.platform.cloud-foundry is not attached to cap.core.type.platform abstract node")},
		},
		"should detect that both Types are not attached to parent node (plural)": {
			knownTypesByPathPattern: nil, // no relations are registered

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					"cap.type.platform.cloud-foundry", "cap.type.platform.nomad", "cap.type.platform.mesos",
				},
			},

			expErrors: []error{errors.New("Types cap.type.platform.cloud-foundry, cap.type.platform.nomad and cap.type.platform.mesos are not attached to cap.core.type.platform abstract node")},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeHubCli := &fakeHub{knownTypesByPathPattern: tc.knownTypesByPathPattern}

			implValidator := manifest.NewRemoteImplementationValidator(fakeHubCli)

			// when
			result, err := implValidator.CheckParentNodesAssociation(context.Background(), tc.relationsToParentNode)

			// then
			require.NoError(t, err)
			assert.Equal(t, result.Errors, tc.expErrors)
		})
	}
}
