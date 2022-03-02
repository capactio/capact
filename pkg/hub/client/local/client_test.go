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

	family, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				Alias:     ptr.String("child"),
				CreatedBy: ptr.String("nature"),
				TypeRef:   typeRef("cap.type.child:0.1.0"),
				Value: map[string]interface{}{
					"name": "Luke Skywalker",
				},
			},
			{
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
				Alias:     ptr.String("parent"),
				CreatedBy: ptr.String("nature"),
				TypeRef:   typeRef("cap.type.complex:0.1.0"),
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
		},
	})
	require.NoError(t, err)

	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: ptr.String("nature"),
	}, WithFields(TypeInstanceAllFields))
	require.NoError(t, err)

	resourcePrinter := cliprinter.NewForResource(os.Stdout, cliprinter.WithTable(typeInstanceDetailsMapper(family, getDataDirectlyFromStorage(t, srvAddr, familyDetails))))
	require.NoError(t, resourcePrinter.Print(familyDetails))

	for _, member := range familyDetails {
		_ = cli.DeleteTypeInstance(ctx, member.ID)
	}
}

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

func getDataDirectlyFromStorage(t *testing.T, addr string, details []gqllocalapi.TypeInstance) map[string]string {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)

	ctx := context.Background()
	client := pb.NewStorageBackendClient(conn)

	var out = map[string]string{}
	for _, ti := range details {
		got, err := client.GetValue(ctx, &pb.GetValueRequest{
			TypeInstanceId:  ti.ID,
			ResourceVersion: 1,
		})
		//require.NoError(t, err)
		if err != nil {
			continue
		}

		out[ti.ID] = string(got.Value)
	}
	return out
}

func typeInstanceDetailsMapper(family []gqllocalapi.CreateTypeInstanceOutput, storage map[string]string) func(inRaw interface{}) (cliprinter.TableData, error) {
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
			out.Headers = []string{"TYPE INSTANCE ID", "ALIAS", "TYPE", "DATA FROM GQL", "BACKEND", "BACKEND CONTEXT", "DATA IN EXTERNAL BACKEND"}
			for _, ti := range in {
				out.MultipleRows = append(out.MultipleRows, []string{
					ti.ID,
					mapping[ti.ID],
					fmt.Sprintf("%s:%s", ti.TypeRef.Path, ti.TypeRef.Revision),
					mustMarshal(ti.LatestResourceVersion.Spec.Value),
					fmt.Sprintf("%s%s", ti.Backend.ID, labelIfAbstract(ti.Backend.Abstract)),
					mustMarshal(ti.LatestResourceVersion.Spec.Backend.Context),
					storage[ti.ID],
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

func typeRef(in string) *gqllocalapi.TypeInstanceTypeReferenceInput {
	out := strings.Split(in, ":")
	return &gqllocalapi.TypeInstanceTypeReferenceInput{Path: out[0], Revision: out[1]}
}
