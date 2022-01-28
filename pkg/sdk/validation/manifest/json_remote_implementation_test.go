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
		knownTypes            []*gqlpublicapi.Type
		relationsToParentNode manifest.ParentNodesAssociation
		expErrors             []error
	}{
		"should success as all nodes are attached to parent nodes": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.nomad", "0.1.0", "cap.core.type.platform"),
			},

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
				},
			},

			expErrors: nil, // no errors
		},
		"should detect that both Type with different revision is not attached to parent node": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.cloud-foundry", "0.2.0", ""),
			},
			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.2.0"},
				},
			},

			expErrors: []error{errors.New("Type cap.type.platform.cloud-foundry:0.2.0 is not attached to cap.core.type.platform abstract node")},
		},
		"should detect that one Type is not attached to parent node (singular)": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.nomad", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", ""),
			},
			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
				},
			},

			expErrors: []error{errors.New("Type cap.type.platform.cloud-foundry:0.1.0 is not attached to cap.core.type.platform abstract node")},
		},
		"should detect that both Types are not attached to parent node (plural)": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.nomad", "0.1.0", ""),
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", ""),
				fixGQLType("cap.type.platform.mesos", "0.1.0", ""),
			},

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
					{Path: "cap.type.platform.mesos", Revision: "0.1.0"},
				},
			},

			expErrors: []error{errors.New("Types cap.type.platform.cloud-foundry:0.1.0, cap.type.platform.nomad:0.1.0 and cap.type.platform.mesos:0.1.0 are not attached to cap.core.type.platform abstract node")},
		},

		"should not report problems with parents for unknown Types": {
			knownTypes: nil, // not Types in Hub

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
					{Path: "cap.type.platform.mesos", Revision: "0.1.0"},
				},
			},

			expErrors: nil, // no error about parents.
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeHubCli := &fakeHub{knownTypes: tc.knownTypes}

			implValidator := manifest.NewRemoteImplementationValidator(fakeHubCli)

			// when
			result, err := implValidator.CheckParentNodesAssociation(context.Background(), tc.relationsToParentNode)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expErrors, result.Errors)
		})
	}
}

func fixGQLType(path, rev, parent string) *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: path,
		Revisions: []*gqlpublicapi.TypeRevision{
			{
				Revision: rev,
				Spec: &gqlpublicapi.TypeSpec{
					AdditionalRefs: []string{parent},
				},
			}},
	}
}
