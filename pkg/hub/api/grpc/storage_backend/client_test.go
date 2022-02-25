package storage_backend_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// #nosec G101
const secretStorageBackendAddr = "GRPC_SECRET_STORAGE_BACKEND_ADDR"

// This test illustrates how to use gRPC Go client against real gRPC Storage Backend server.
//
// NOTE: Before running this test, make sure that the server is running under the `srvAddr` address
//	and the `provider` is enabled on this server.
//
// To run this test, execute:
// GRPC_SECRET_STORAGE_BACKEND_ADDR=":50051" go test ./pkg/hub/api/grpc/storage_backend -v
func TestNewStorageBackendClient(t *testing.T) {
	srvAddr := os.Getenv(secretStorageBackendAddr)
	if srvAddr == "" {
		t.Skipf("skipping storage backend gRPC client test as the env %s is not provided", secretStorageBackendAddr)
	}
	provider := "dotenv"

	valueBytes := []byte(`{"key": true}`)
	typeInstanceID := "id"

	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, provider))

	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	require.NoError(t, err)

	ctx := context.Background()
	client := pb.NewStorageBackendClient(conn)

	// create
	t.Logf("Creating TI %q...\n", typeInstanceID)

	_, err = client.OnCreate(ctx, &pb.OnCreateRequest{
		TypeinstanceId: typeInstanceID,
		Value:          valueBytes,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// get value

	var resourceVersion uint32 = 1
	res, err := client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, valueBytes, res.Value)

	// update
	t.Logf("Updating TI %q...\n", typeInstanceID)

	newValueBytes := []byte(`{"key": "updated"}`)
	_, err = client.OnUpdate(ctx, &pb.OnUpdateRequest{
		TypeinstanceId:     typeInstanceID,
		NewResourceVersion: 2,
		NewValue:           newValueBytes,
		Context:            reqContext,
	})
	require.NoError(t, err)

	// get value

	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, valueBytes, res.Value)

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, newValueBytes, res.Value)

	// lock

	t.Logf("Locking TI %q...\n", typeInstanceID)

	_, err = client.OnLock(ctx, &pb.OnLockRequest{
		TypeinstanceId: typeInstanceID,
		Context:        reqContext,
		LockedBy:       "test/sample",
	})
	require.NoError(t, err)

	// get lockedBy

	lockedByRes, err := client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeinstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	require.NotNil(t, lockedByRes.LockedBy)
	assert.Equal(t, "test/sample", *lockedByRes.LockedBy)
	t.Logf("Getting TI %q: locked by %q\n", typeInstanceID, *lockedByRes.LockedBy)

	// unlock

	t.Logf("Unlocking TI %q...\n", typeInstanceID)

	_, err = client.OnUnlock(ctx, &pb.OnUnlockRequest{
		TypeinstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// get lockedBy

	lockedByRes, err = client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeinstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: locked by: %v\n", typeInstanceID, lockedByRes.LockedBy)
	assert.Nil(t, lockedByRes.LockedBy)

	// get value

	resourceVersion = 1
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, valueBytes, res.Value)

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.NoError(t, err)

	t.Logf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))
	assert.Equal(t, newValueBytes, res.Value)

	// delete
	t.Logf("Deleting TI %q...\n", typeInstanceID)

	_, err = client.OnDelete(ctx, &pb.OnDeleteRequest{
		TypeinstanceId: typeInstanceID,
		Context:        reqContext,
	})
	require.NoError(t, err)

	// last get

	resourceVersion = 1
	_, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:  typeInstanceID,
		ResourceVersion: resourceVersion,
		Context:         reqContext,
	})
	require.Error(t, err)
	assert.EqualError(t, err, "rpc error: code = NotFound desc = TypeInstance \"id\" in revision 1 was not found")
	t.Logf("Getting TI %q: resource version %d: error: %v\n", typeInstanceID, resourceVersion, err)
}
