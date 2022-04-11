//go:build localhub
// +build localhub

package localhub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"capact.io/capact/pkg/hub/client/local"
	"gotest.tools/assert"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const storageBackendAddr = "GRPC_SECRET_STORAGE_BACKEND_ADDR"

type expectedTypeInstanceData struct {
	alias                   *string
	backendContext          interface{}
	dataInGQL               interface{}
	dataInExternalBackend   interface{}
	locked                  string
	lockedInExternalBackend string
}

// This test:
// - register TypeInstance for external Storage Backend
// - create a new TypeInstance stored in built-in backend
// - create a new TypeInstance stored in registered external backend
// - create a new TypeInstance stored in registered external backend with custom context
// - update TypeInstances
// - lock/unlock TypeInstance
// - delete all created TypeInstance
func TestExternalStorage(t *testing.T) {
	srvAddr := os.Getenv(storageBackendAddr)
	if srvAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", storageBackendAddr)
	}

	ctx := context.Background()
	cli := getLocalClient(t)
	dotenvHubStorage, cleanup := registerExternalDotenvStorage(ctx, t, cli, srvAddr)
	defer cleanup(t)

	t.Log("Create TypeInstances")
	inputTypeInstances := []*gqllocalapi.CreateTypeInstanceInput{
		{
			// This TypeInstance:
			// - is stored in built-in backend
			// - doesn't have backend context
			Alias:     ptr.String("child"),
			CreatedBy: ptr.String("nature"),
			TypeRef:   typeRef("cap.type.child:0.1.0"),
			Value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
		},
		{
			// This TypeInstance:
			// - is stored in external backend
			// - has additional context
			Alias:     ptr.String("second-child"),
			CreatedBy: ptr.String("nature"),
			TypeRef:   typeRef("cap.type.child:0.2.0"),
			Value: map[string]interface{}{
				"name": "Leia Organa",
			},
			Backend: &gqllocalapi.TypeInstanceBackendInput{
				ID: dotenvHubStorage.ID,
				Context: map[string]interface{}{
					"provider": "dotenv",
				},
			},
		},
		{
			// This TypeInstance:
			// - is stored in external backend
			// - doesn't have additional context
			Alias:     ptr.String("original"),
			CreatedBy: ptr.String("nature"),
			TypeRef:   typeRef("cap.type.original:0.2.0"),
			Value: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			Backend: &gqllocalapi.TypeInstanceBackendInput{
				ID: dotenvHubStorage.ID,
				// no context
			},
		},
		{
			// This TypeInstance:
			// - is stored in external backend
			// - has additional context
			Alias:     ptr.String("parent"),
			CreatedBy: ptr.String("nature"),
			TypeRef:   typeRef("cap.type.parent:0.1.0"),
			Value: map[string]interface{}{
				"name": "Darth Vader",
			},
			Backend: &gqllocalapi.TypeInstanceBackendInput{
				ID: dotenvHubStorage.ID,
				Context: map[string]interface{}{
					"provider": "dotenv",
				},
			},
		},
	}
	family, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: inputTypeInstances,
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
			{From: "parent", To: "child"},
			{From: "parent", To: "second-child"},
			{From: "parent", To: "original"},
		},
	})
	require.NoError(t, err)

	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, local.WithFields(local.TypeInstanceAllFields))
	require.NoError(t, err)

	defer removeAllMembers(t, cli, familyDetails)
	assertTypeInstancesDetail(t, familyDetails, family, getDataDirectlyFromStorage(t, srvAddr, familyDetails), []*expectedTypeInstanceData{
		{
			alias: ptr.String("second-child"),
			backendContext: map[string]interface{}{
				"provider": "dotenv",
			},
			dataInGQL: map[string]interface{}{
				"name": "Leia Organa",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Leia Organa",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias:          ptr.String("child"),
			backendContext: nil,
			dataInGQL: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			dataInExternalBackend:   nil,
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias:          ptr.String("original"),
			backendContext: nil,
			dataInGQL: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias: ptr.String("parent"),
			backendContext: map[string]interface{}{
				"provider": "dotenv",
			},
			dataInGQL: map[string]interface{}{
				"name": "Darth Vader",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Darth Vader",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
	})

	t.Log("Update TypeInstances")
	// - for cap.type.parent don't update `value` and `context` - use old ones
	// - for cap.type.original:0.2.0 don't update `value`, and zero the `context`
	// - for all others, update both `value` and `context`
	toUpdate := make([]gqllocalapi.UpdateTypeInstancesInput, 0, len(familyDetails))
	for idx, member := range familyDetails {
		val := map[string]interface{}{
			"updated-value": fmt.Sprintf("context %d", idx),
		}
		backend := &gqllocalapi.UpdateTypeInstanceBackendInput{
			Context: map[string]interface{}{
				"updated": fmt.Sprintf("context %d", idx),
			},
		}
		// For cap.type.original:0.2.0 don't update value, and zero the `context`.
		if member.TypeRef.Path == "cap.type.original" {
			val = nil
			backend.Context = nil
		}

		// 	// For cap.type.parent don't update value and `context` - use old ones
		if member.TypeRef.Path == "cap.type.parent" {
			val = nil
			backend = nil
		}
		toUpdate = append(toUpdate, gqllocalapi.UpdateTypeInstancesInput{
			ID:        member.ID,
			CreatedBy: ptr.String("update"),
			TypeInstance: &gqllocalapi.UpdateTypeInstanceInput{
				Value:   val,
				Backend: backend,
			},
		})
	}

	updatedFamily, err := cli.UpdateTypeInstances(ctx, toUpdate)
	require.NoError(t, err)
	assertTypeInstancesDetail(t, updatedFamily, family, getDataDirectlyFromStorage(t, srvAddr, updatedFamily), []*expectedTypeInstanceData{
		{
			alias: ptr.String("second-child"),
			backendContext: map[string]interface{}{
				"updated": "context 0",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 0",
			},
			dataInExternalBackend: map[string]interface{}{
				"updated-value": "context 0",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias: ptr.String("child"),
			backendContext: map[string]interface{}{
				"updated": "context 1",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 1",
			},
			dataInExternalBackend:   nil,
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias:          ptr.String("original"),
			backendContext: nil,
			dataInGQL: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias: ptr.String("parent"),
			backendContext: map[string]interface{}{
				"provider": "dotenv",
			},
			dataInGQL: map[string]interface{}{
				"name": "Darth Vader",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Darth Vader",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
	})

	t.Log("Locking TypeInstances")
	var ids []string
	for _, member := range familyDetails {
		ids = append(ids, member.ID)
	}
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     ids,
		OwnerID: "demo/testing",
	})
	require.NoError(t, err)
	familyDetails, err = cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, local.WithFields(local.TypeInstanceAllFields))
	require.NoError(t, err)

	assertTypeInstancesDetail(t, familyDetails, family, getDataDirectlyFromStorage(t, srvAddr, familyDetails), []*expectedTypeInstanceData{
		{
			alias: ptr.String("second-child"),
			backendContext: map[string]interface{}{
				"updated": "context 0",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 0",
			},
			dataInExternalBackend: map[string]interface{}{
				"updated-value": "context 0",
			},
			locked:                  "demo/testing",
			lockedInExternalBackend: "demo/testing",
		},
		{
			alias: ptr.String("child"),
			backendContext: map[string]interface{}{
				"updated": "context 1",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 1",
			},
			dataInExternalBackend:   nil,
			locked:                  "demo/testing",
			lockedInExternalBackend: "",
		},
		{
			alias:          ptr.String("original"),
			backendContext: nil,
			dataInGQL: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			locked:                  "demo/testing",
			lockedInExternalBackend: "demo/testing",
		},
		{
			alias: ptr.String("parent"),
			backendContext: map[string]interface{}{
				"provider": "dotenv",
			},
			dataInGQL: map[string]interface{}{
				"name": "Darth Vader",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Darth Vader",
			},
			locked:                  "demo/testing",
			lockedInExternalBackend: "demo/testing",
		},
	})

	t.Log("Unlocking TypeInstances")
	err = cli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		Ids:     ids,
		OwnerID: "demo/testing",
	})
	require.NoError(t, err)
	familyDetails, err = cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, local.WithFields(local.TypeInstanceAllFields))
	require.NoError(t, err)

	assertTypeInstancesDetail(t, familyDetails, family, getDataDirectlyFromStorage(t, srvAddr, familyDetails), []*expectedTypeInstanceData{
		{
			alias: ptr.String("second-child"),
			backendContext: map[string]interface{}{
				"updated": "context 0",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 0",
			},
			dataInExternalBackend: map[string]interface{}{
				"updated-value": "context 0",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias: ptr.String("child"),
			backendContext: map[string]interface{}{
				"updated": "context 1",
			},
			dataInGQL: map[string]interface{}{
				"updated-value": "context 1",
			},
			dataInExternalBackend:   nil,
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias:          ptr.String("original"),
			backendContext: nil,
			dataInGQL: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Anakin Skywalke",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
		{
			alias: ptr.String("parent"),
			backendContext: map[string]interface{}{
				"provider": "dotenv",
			},
			dataInGQL: map[string]interface{}{
				"name": "Darth Vader",
			},
			dataInExternalBackend: map[string]interface{}{
				"name": "Darth Vader",
			},
			locked:                  "",
			lockedInExternalBackend: "",
		},
	})
}

func registerExternalDotenvStorage(ctx context.Context, t *testing.T, cli *local.Client, srvAddr string) (gqllocalapi.CreateTypeInstanceOutput, func(t *testing.T)) {
	t.Helper()

	ti, err := cli.CreateTypeInstances(ctx, fixExternalDotenvStorage(t, srvAddr))
	require.NoError(t, err)
	require.Len(t, ti, 1)
	dotenvHubStorage := ti[0]

	return dotenvHubStorage, func(t *testing.T) {
		err = cli.DeleteTypeInstance(ctx, dotenvHubStorage.ID)
		require.NoError(t, err)
	}
}

func unmarshalContextSchema(t *testing.T, schema string) interface{} {
	var contextSchema interface{}
	err := json.Unmarshal([]byte(schema), &contextSchema)
	require.NoError(t, err)
	return contextSchema
}

func fixExternalDotenvStorage(t *testing.T, addr string) *gqllocalapi.CreateTypeInstancesInput {
	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				CreatedBy: ptr.String("manually"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.example.filesystem.storage",
					Revision: "0.1.0",
				},
				Value: storageSpec{
					URL:         ptr.String(addr),
					AcceptValue: ptr.Bool(true),
					ContextSchema: unmarshalContextSchema(t, heredoc.Doc(`
				      {
				      	"$id": "#/properties/contextSchema",
				      	"type": "object",
				      	"properties": {
				      		"provider": {
				      			"$id": "#/properties/contextSchema/properties/name",
				      			"type": "string",
								  "enum": [
									"aws_secretsmanager",
									"dotenv"
								  ]
				      		}
				      	},
				      	"additionalProperties": true
				      }`)),
				},
			},
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{},
	}
}

type externalData struct {
	Value    string
	LockedBy *string
}

func getDataDirectlyFromStorage(t *testing.T, addr string, details []gqllocalapi.TypeInstance) map[string]externalData {
	t.Helper()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)

	ctx := context.Background()
	client := pb.NewValueAndContextStorageBackendClient(conn)

	var out = map[string]externalData{}
	for _, ti := range details {
		val, err := client.GetValue(ctx, &pb.GetValueRequest{
			TypeInstanceId:  ti.ID,
			ResourceVersion: uint32(ti.LatestResourceVersion.ResourceVersion),
		})
		if err != nil {
			t.Logf("error while getting value from storage for TypeInstance %s", ti.ID)
			continue
		}

		locked, err := client.GetLockedBy(ctx, &pb.GetLockedByRequest{
			TypeInstanceId: ti.ID,
		})
		require.NoError(t, err)

		out[ti.ID] = externalData{
			Value:    string(val.Value),
			LockedBy: locked.LockedBy,
		}
	}
	return out
}

func assertTypeInstancesDetail(t *testing.T, typeInstances interface{}, family []gqllocalapi.CreateTypeInstanceOutput, storage map[string]externalData, expectedData []*expectedTypeInstanceData) {
	mapping := map[string]string{}
	for _, member := range family {
		mapping[member.ID] = member.Alias
	}

	switch in := typeInstances.(type) {
	case []gqllocalapi.TypeInstance:
		for _, ti := range in {
			expectedTI, err := findExpectedTypeInstance(expectedData, mapping[ti.ID])
			require.NoError(t, err)
			dataInExternalBackend := mustMarshal(t, expectedTI.dataInExternalBackend)
			if dataInExternalBackend == "null" {
				assert.Equal(t, storage[ti.ID].Value, "")
			} else {
				assert.Equal(t, storage[ti.ID].Value, dataInExternalBackend)
			}
			assert.Equal(t, mustMarshal(t, ti.LatestResourceVersion.Spec.Backend.Context), mustMarshal(t, expectedTI.backendContext))
			assert.Equal(t, mustMarshal(t, ti.LatestResourceVersion.Spec.Value), mustMarshal(t, expectedTI.dataInGQL))
			assert.Equal(t, stringDefault(ti.LockedBy, ""), expectedTI.locked)
			assert.Equal(t, stringDefault(storage[ti.ID].LockedBy, ""), expectedTI.lockedInExternalBackend)
		}
	}
}

func findExpectedTypeInstance(typeInstances []*expectedTypeInstanceData, alias string) (*expectedTypeInstanceData, error) {
	for _, ti := range typeInstances {
		if *ti.alias == alias {
			return ti, nil
		}
	}
	return nil, fmt.Errorf("cannot find TypeInstance with alias %s", alias)
}

func mustMarshal(t *testing.T, v interface{}) string {
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return string(out)
}

func stringDefault(in *string, def string) string {
	if in == nil {
		return def
	}
	return *in
}

func removeAllMembers(t *testing.T, cli *local.Client, familyDetails []gqllocalapi.TypeInstance) {
	t.Helper()

	ctx := context.Background()
	for _, member := range familyDetails {
		if member.TypeRef.Path != "cap.type.parent" {
			defer func(id string) { // delay the child deletions
				t.Logf("Delete child %s", id)
				err := cli.DeleteTypeInstance(ctx, id)
				require.NoError(t, err)
			}(member.ID)

			continue
		}

		t.Logf("Delete parent %v", member.ID)

		// Delete parent first, to unblock deletion of children
		err := cli.DeleteTypeInstance(ctx, member.ID)
		require.NoError(t, err)
	}
}
