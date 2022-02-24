package storage_backend_test

import (
	"context"
	"fmt"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"google.golang.org/grpc"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

// This example illustrates how to use gRPC Go client against real gRPC Storage Backend server.
//
// NOTE: Before running this example, make sure that the server is running under the `srvAddr` address
//	and the `provider` is enabled on this server.
func ExampleNewStorageBackendClient() {
	provider := "dotenv"
	srvAddr := ":50051" // server address

	valueBytes := []byte(`{"key": true}`)
	typeInstanceID := utilrand.String(10) // temp

	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, provider))

	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := pb.NewStorageBackendClient(conn)

	// create
	fmt.Printf("Creating TI %q...\n", typeInstanceID)

	_, err = client.OnCreate(ctx, &pb.OnCreateRequest{
		TypeinstanceId:       typeInstanceID,
		Value:                valueBytes,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// get value

	var resourceVersion uint32 = 1
	res, err := client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))

	// update
	fmt.Printf("Updating TI %q...\n", typeInstanceID)

	newValueBytes := []byte(`{"key": "updated"}`)
	_, err = client.OnUpdate(ctx, &pb.OnUpdateRequest{
		TypeinstanceId:       typeInstanceID,
		NewResourceVersion:   2,
		NewValue:             newValueBytes,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// get value

	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))

	// lock

	fmt.Printf("Locking TI %q...\n", typeInstanceID)

	_, err = client.OnLock(ctx, &pb.OnLockRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
		LockedBy:             "test/sample",
	})
	if err != nil {
		panic(err)
	}

	// get lockedBy

	lockedByRes, err := client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	if lockedByRes.LockedBy == nil {
		panic("lockedBy cannot be nil")
	}

	fmt.Printf("Getting TI %q: locked by %q\n", typeInstanceID, *lockedByRes.LockedBy)

	// unlock

	fmt.Printf("Unlocking TI %q...\n", typeInstanceID)

	_, err = client.OnUnlock(ctx, &pb.OnUnlockRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// get lockedBy

	lockedByRes, err = client.GetLockedBy(ctx, &pb.GetLockedByRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: locked by: %v\n", typeInstanceID, lockedByRes.LockedBy)

	// get value

	resourceVersion = 1
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))

	resourceVersion = 2
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: %s\n", typeInstanceID, resourceVersion, string(res.Value))

	// delete
	fmt.Printf("Deleting TI %q...\n", typeInstanceID)

	_, err = client.OnDelete(ctx, &pb.OnDeleteRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// last get

	resourceVersion = 1
	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      resourceVersion,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Getting TI %q: resource version %d: is nil: %v\n", typeInstanceID, resourceVersion, res.Value == nil)
}
