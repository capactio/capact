package storage_backend_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

// #nosec G101
const secretStorageBackendAddrEnv = "GRPC_SECRET_STORAGE_BACKEND_ADDR"
const typeInstanceIDEnv = "TYPEINSTANCE_ID"

// This test illustrates how to use gRPC Go client against real gRPC Storage Backend server.
//
// NOTE: Before running this test, make sure that the server is running under the `srvAddr` address
//	and the `provider` is enabled on this server.
//
// To run this test, execute:
// GRPC_SECRET_STORAGE_BACKEND_ADDR=":50051" go test ./pkg/hub/api/grpc/storage_backend -run "^TestNewStorageBackendClient$" -v -count 1
//
// You can run this test with custom TypeInstance ID, by setting TYPEINSTANCE_ID env variable during test run.
// This might be helpful while running this test against server with different default provider configured.
func TestNewStorageBackendClient(t *testing.T) {
	srvAddr := os.Getenv(secretStorageBackendAddrEnv)
	if srvAddr == "" {
		t.Skipf("skipping storage backend gRPC client test as the env %s is not provided", secretStorageBackendAddrEnv)
	}

	value := []byte(`{"key": true}`)
	typeInstanceID := os.Getenv(typeInstanceIDEnv)
	if typeInstanceID == "" {
		// fallback to default
		typeInstanceID = "id"
	}
	provider := "dotenv"
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, provider))

	executeSecretStorageBackendTestScenario(t, srvAddr, typeInstanceID, value, reqContext)
}

// This test illustrates how to use gRPC Go client against real gRPC Storage Backend server without passing request context.
//
// NOTE: Before running this test, make sure that the server is running under the `srvAddr` address
//	and there is just one `provider` enabled on this server.
//
// To run this test, execute:
// GRPC_SECRET_STORAGE_BACKEND_ADDR=":50051" go test ./pkg/hub/api/grpc/storage_backend -run "^TestNewStorageBackendClient_WithDefaultProvider$" -v -count 1
//
// You can run this test with custom TypeInstance ID, by setting TYPEINSTANCE_ID env variable during test run.
// This might be helpful while running this test against server with different default provider configured.
func TestNewStorageBackendClient_WithDefaultProvider(t *testing.T) {
	srvAddr := os.Getenv(secretStorageBackendAddrEnv)
	if srvAddr == "" {
		t.Skipf("skipping storage backend gRPC client test as the env %s is not provided", secretStorageBackendAddrEnv)
	}

	typeInstanceID := os.Getenv(typeInstanceIDEnv)
	if typeInstanceID == "" {
		// fallback to default
		typeInstanceID = "id"
	}

	value := []byte(`{"key": true}`)

	executeSecretStorageBackendTestScenario(t, srvAddr, typeInstanceID, value, nil)
}

func executeSecretStorageBackendTestScenario(t *testing.T, srvAddr, typeInstanceID string, value, reqContext []byte) {
	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	require.NoError(t, err)

	ctx := context.Background()
	client := pb.NewValueAndContextStorageBackendClient(conn)

	// create
	t.Logf("Creating TI %q...\n", typeInstanceID)

	_, err = client.OnCreate(ctx, &pb.OnCreateValueAndContextRequest{
		TypeInstanceId: typeInstanceID,
		Value:          value,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// get value

	var resourceVersion uint32 = 1
	res, err := client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, value, res.Value)

	// update
	t.Logf("Updating TI %q...\n", typeInstanceID)

	newValueBytes := []byte(`{"key": "updated"}`)
	_, err = client.OnUpdate(ctx, &pb.OnUpdateValueAndContextRequest{
		TypeInstanceId:     typeInstanceID,
		NewResourceVersion: 2,
		NewValue:           newValueBytes,
		Context:            reqContext,
	})
	require.NoError(t, err)

	// get value

	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, value, res.Value)

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, newValueBytes, res.Value)

	// lock

	t.Logf("Locking TI %q...\n", typeInstanceID)

	_, err = client.OnLock(ctx, &pb.OnLockRequest{
		TypeInstanceId: typeInstanceID,
		Context:        reqContext,
		LockedBy:       "test/sample",
	})
	require.NoError(t, err)

	// get lockedBy

	lockedByRes, err := client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeInstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	require.NotNil(t, lockedByRes.LockedBy)
	assert.Equal(t, "test/sample", *lockedByRes.LockedBy)
	t.Logf("Getting TI %q: locked by %q\n", typeInstanceID, *lockedByRes.LockedBy)

	// unlock

	t.Logf("Unlocking TI %q...\n", typeInstanceID)

	_, err = client.OnUnlock(ctx, &pb.OnUnlockRequest{
		TypeInstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// get lockedBy

	lockedByRes, err = client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeInstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: locked by: %v\n", typeInstanceID, lockedByRes.LockedBy)
	assert.Nil(t, lockedByRes.LockedBy)

	// get value

	resourceVersion = 1
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, value, res.Value)

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, newValueBytes, res.Value)

	// delete second revision
	t.Logf("Deleting TI %q revision %d...\n", typeInstanceID, resourceVersion)
	_, err = client.OnDeleteRevision(ctx, &pb.OnDeleteRevisionRequest{
		TypeInstanceId:  typeInstanceID,
		Context:         reqContext,
		ResourceVersion: resourceVersion,
	})
	require.NoError(t, err)

	_, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf(`rpc error: code = NotFound desc = TypeInstance "%s" in revision 2 was not found`, typeInstanceID))

	t.Logf("Getting TI %q: resource version %d: error: %v\n", typeInstanceID, resourceVersion, err)

	// delete
	t.Logf("Deleting the whole TI %q...\n", typeInstanceID)

	_, err = client.OnDelete(ctx, &pb.OnDeleteValueAndContextRequest{
		TypeInstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// last get

	resourceVersion = 1
	_, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("rpc error: code = NotFound desc = TypeInstance \"%s\" in revision 1 was not found", typeInstanceID))
	t.Logf("Getting TI %q: resource version %d: error: %v\n", typeInstanceID, resourceVersion, err)
}
