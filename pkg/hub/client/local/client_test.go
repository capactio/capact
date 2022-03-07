package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	cliprinter "capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// TODO(review): THIS FILE WILL BE REMOVED BEFORE MERGING. IT WAS ADDED ONLY FOR DEMO/PR TESTING PURPOSES.

// #nosec G101
const secretStorageBackendAddr = "GRPC_SECRET_STORAGE_BACKEND_ADDR"

type StorageValue struct {
	URL           string `json:"url"`
	AcceptValue   bool   `json:"acceptValue"`
	ContextSchema string `json:"contextSchema"`
}

// This test showcase how to use GraphQL client to:
// - register TypeInstance for external Storage Backend
// - create a new TypeInstance stored in built-in backend
// - create a new TypeInstance stored in registered external backend
// - create a new TypeInstance stored in registered external backend with custom context
// - update TypeInstances
// - lock/unlock TypeInstance
// - delete all created TypeInstance
//
// Prerequisite:
//   Before running this test, make sure that the external backend is running:
//     APP_LOGGER_DEV_MODE=true APP_SUPPORTED_PROVIDERS="dotenv" go run ./cmd/secret-storage-backend/main.go
//   and Local Hub:
//     cd hub-js; APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_HUB_MODE=local npm run dev; cd ..
//
// To run this test, execute:
// GRPC_SECRET_STORAGE_BACKEND_ADDR="0.0.0.0:50051" go test ./pkg/hub/client/local/ -v -count 1
func TestThatShowcaseExternalStorage(t *testing.T) {
	srvAddr := os.Getenv(secretStorageBackendAddr)
	if srvAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", secretStorageBackendAddr)
	}

	ctx := context.Background()
	cli := NewDefaultClient("http://localhost:8080/graphql")
	dotenvHubStorage, cleanup := registerExternalDotenvStorage(ctx, t, cli, srvAddr)
	defer cleanup()

	// SCENARIO - CREATE
	family, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
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
				// - should be stored with mutated context (from create req)
				Alias:     ptr.String("second-child"),
				CreatedBy: ptr.String("nature"),
				TypeRef:   typeRef("cap.type.child:0.2.0"),
				Value: map[string]interface{}{
					"name": "Leia Organa",
				},
				Backend: &gqllocalapi.TypeInstanceBackendInput{
					ID: dotenvHubStorage.ID,
					Context: map[string]interface{}{
						"provider": "mock-me", // this will inform external backend to return mutated context
					},
				},
			},
			{
				// This TypeInstance:
				// - is stored in external backend
				// - doesn't have additional context
				// - should be stored without mutated context
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
				// - should be stored without mutated context
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
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
			{From: "parent", To: "child"},
			{From: "parent", To: "second-child"},
			{From: "parent", To: "original"},
		},
	})
	require.NoError(t, err)

	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, WithFields(TypeInstanceAllFields))
	require.NoError(t, err)

	defer removeAllMembers(t, cli, familyDetails)

	fmt.Print("\n\n======== After create result  ============\n\n")
	resourcePrinter := cliprinter.NewForResource(os.Stdout, cliprinter.WithTable(typeInstanceDetailsMapper(family, getDataDirectlyFromStorage(t, srvAddr, familyDetails))))
	require.NoError(t, resourcePrinter.Print(familyDetails))

	// SCENARIO - UPDATE
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

		// For cap.type.parent don't update value and `context` - use old ones
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

	fmt.Print("\n\n======== After update result  ============\n\n")
	resourcePrinter = cliprinter.NewForResource(os.Stdout, cliprinter.WithTable(typeInstanceDetailsMapper(family, getDataDirectlyFromStorage(t, srvAddr, updatedFamily))))
	require.NoError(t, resourcePrinter.Print(updatedFamily))

	// SCENARIO - LOCK
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
	}, WithFields(TypeInstanceAllFields))
	require.NoError(t, err)

	fmt.Print("\n\n======== After locking result  ============\n\n")
	resourcePrinter = cliprinter.NewForResource(os.Stdout, cliprinter.WithTable(typeInstanceDetailsMapper(family, getDataDirectlyFromStorage(t, srvAddr, familyDetails))))
	require.NoError(t, resourcePrinter.Print(familyDetails))

	// SCENARIO - UNLOCK
	err = cli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		Ids:     ids,
		OwnerID: "demo/testing",
	})
	require.NoError(t, err)
	familyDetails, err = cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, WithFields(TypeInstanceAllFields))
	require.NoError(t, err)

	fmt.Print("\n\n======== After unlocking result  ============\n\n")
	resourcePrinter = cliprinter.NewForResource(os.Stdout, cliprinter.WithTable(typeInstanceDetailsMapper(family, getDataDirectlyFromStorage(t, srvAddr, familyDetails))))
	require.NoError(t, resourcePrinter.Print(familyDetails))
}

// ======= HELPERS =======

func registerExternalDotenvStorage(ctx context.Context, t *testing.T, cli *Client, srvAddr string) (gqllocalapi.CreateTypeInstanceOutput, func()) {
	t.Helper()

	ti, err := cli.CreateTypeInstances(ctx, fixExternalDotenvStorage(srvAddr))
	require.NoError(t, err)
	require.Len(t, ti, 1)
	dotenvHubStorage := ti[0]

	return dotenvHubStorage, func() {
		_ = cli.DeleteTypeInstance(ctx, dotenvHubStorage.ID)
	}
}

func fixExternalDotenvStorage(addr string) *gqllocalapi.CreateTypeInstancesInput {
	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				CreatedBy: ptr.String("manually"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.example.filesystem.storage",
					Revision: "0.1.0",
				},
				Value: StorageValue{
					URL:         addr,
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
	client := pb.NewStorageBackendClient(conn)

	var out = map[string]externalData{}
	for _, ti := range details {
		val, err := client.GetValue(ctx, &pb.GetValueRequest{
			TypeInstanceId:  ti.ID,
			ResourceVersion: uint32(ti.LatestResourceVersion.ResourceVersion),
		})
		if err != nil {
			continue
		}

		locked, err := client.GetLockedBy(ctx, &pb.GetLockedByRequest{
			TypeInstanceId: ti.ID,
		})
		if err != nil {
			continue
		}

		out[ti.ID] = externalData{
			Value:    string(val.Value),
			LockedBy: locked.LockedBy,
		}
	}
	return out
}

func typeInstanceDetailsMapper(family []gqllocalapi.CreateTypeInstanceOutput, storage map[string]externalData) func(inRaw interface{}) (cliprinter.TableData, error) {
	mapping := map[string]string{}
	for _, member := range family {
		mapping[member.ID] = member.Alias
	}
	labelIfAbstract := func(in bool) string {
		if in {
			return " (abstract)"
		}
		return ""
	}
	return func(inRaw interface{}) (cliprinter.TableData, error) {
		out := cliprinter.TableData{}

		switch in := inRaw.(type) {
		case []gqllocalapi.TypeInstance:
			out.Headers = []string{"TYPE INSTANCE ID", "ALIAS", "TYPE", "BACKEND", "BACKEND CONTEXT", "DATA IN GQL", "DATA IN exBACKEND", "LOCKED", "LOCKED IN exBACKEND"}
			for _, ti := range in {
				out.MultipleRows = append(out.MultipleRows, []string{
					ti.ID,
					mapping[ti.ID],
					fmt.Sprintf("%s:%s", ti.TypeRef.Path, ti.TypeRef.Revision),
					fmt.Sprintf("%s%s", ti.Backend.ID, labelIfAbstract(ti.Backend.Abstract)),
					mustMarshal(ti.LatestResourceVersion.Spec.Backend.Context),
					mustMarshal(ti.LatestResourceVersion.Spec.Value),
					storage[ti.ID].Value,
					stringDefault(ti.LockedBy, "-"),
					stringDefault(storage[ti.ID].LockedBy, "-"),
				})
			}
		default:
			return cliprinter.TableData{}, fmt.Errorf("got unexpected input type, expected []gqllocalapi.TypeInstance, got %T", inRaw)
		}

		return out, nil
	}
}

func mustMarshal(v interface{}) string {
	out, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func stringDefault(in *string, def string) string {
	if in == nil {
		return def
	}
	return *in
}

func typeRef(in string) *gqllocalapi.TypeInstanceTypeReferenceInput {
	out := strings.Split(in, ":")
	return &gqllocalapi.TypeInstanceTypeReferenceInput{Path: out[0], Revision: out[1]}
}

func removeAllMembers(t *testing.T, cli *Client, familyDetails []gqllocalapi.TypeInstance) {
	t.Helper()

	ctx := context.Background()

	for _, member := range familyDetails {
		if member.TypeRef.Path != "cap.type.parent" {
			defer func(id string) { // delay the child deletions
				fmt.Println("Delete child", id)
				err := cli.DeleteTypeInstance(ctx, id)
				if err != nil {
					t.Logf("err for %v: %v", id, err)
				}
			}(member.ID)

			continue
		}

		fmt.Println("Delete parent", member.ID)

		// Delete parent first, to unblock deletion of children
		err := cli.DeleteTypeInstance(ctx, member.ID)
		if err != nil {
			t.Logf("err for %v: %v", member.ID, err)
		}
	}
}
