package local

import (
	"context"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/require"
)

type StorageValue struct {
	URL           string `json:"url"`
	AcceptValue   bool   `json:"acceptValue"`
	ContextSchema string `json:"contextSchema"`
}

func Test(t *testing.T) {
	ctx := context.Background()
	cli := NewDefaultClient("http://localhost:8080/graphql")
	ti, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				CreatedBy: ptr.String("manually"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.aws.secret-manager.storage",
					Revision: "0.1.0",
				},
				Value: StorageValue{
					URL:         "0.0.0.0:50051",
					AcceptValue: true,
					ContextSchema: heredoc.Doc(`
				      {
				      	"$id": "#/properties/contextSchema",
				      	"type": "object",
				      	"properties": {
				      		"provider": {
				      			"$id": "#/properties/contextSchema/properties/name",
				      			"type": "string",
				      			"const": "dotenv"
				      		}
				      	},
				      	"additionalProperties": false
				      }`),
				},
			},
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{},
	})
	require.NoError(t, err)
	require.Len(t, ti, 1)
	dotenvHubStorage := ti[0]
	defer cli.DeleteTypeInstance(ctx, dotenvHubStorage.ID)

	family, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				CreatedBy: ptr.String("nature"),
				Alias:     ptr.String("child"),
				TypeRef:   typeRef("cap.type.simple:0.1.0"),
				Value: map[string]interface{}{
					"name": "Luke Skywalker",
				},
			},
			{
				CreatedBy: ptr.String("nature"),
				Alias:     ptr.String("parent"),
				TypeRef:   typeRef("cap.type.complex:0.1.0"),
				Value: map[string]interface{}{
					"name": "Darth Vader",
				},
				Backend: &gqllocalapi.TypeInstanceBackendInput{
					ID: dotenvHubStorage.ID,
				},
			},
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
			{From: "parent", To: "child"},
		},
	})
	require.NoError(t, err)

	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, WithFields(TypeInstanceAllFields))

	litter.Dump(familyDetails)
	for _, member := range family {
		require.NoError(t, cli.DeleteTypeInstance(ctx, member.ID))
	}
}

func typeRef(in string) *gqllocalapi.TypeInstanceTypeReferenceInput {
	out := strings.Split(in, ":")
	return &gqllocalapi.TypeInstanceTypeReferenceInput{Path: out[0], Revision: out[1]}
}
