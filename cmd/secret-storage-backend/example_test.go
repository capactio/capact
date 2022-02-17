package main_test

import (
	"context"
	"fmt"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"testing"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"google.golang.org/grpc"
)

func Test(t *testing.T) {
	valueBytes := []byte(`{"key": true}`)
	typeInstanceID := utilrand.String(10) // temp

	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"dotenv"}`))

	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := pb.NewStorageBackendClient(conn)

	// create
	fmt.Println("create TI", typeInstanceID)

	_, err = client.OnCreate(ctx, &pb.OnCreateRequest{
		TypeinstanceId:       typeInstanceID,
		Value:                valueBytes,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// get value

	res, err := client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      1,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("first get - resource version 1", string(res.Value))

	// update
	fmt.Println("update TI", typeInstanceID)

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
		ResourceVersion:      1,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("get after update - resource version 1", string(res.Value))

	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      2,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("get after update - resource version 2", string(res.Value))

	// lock

	fmt.Println("locking")

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

	fmt.Println("first get - lockedBy", *lockedByRes.LockedBy)

	// unlock

	fmt.Println("unlocking")

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

	fmt.Println("second get - lockedBy", lockedByRes.LockedBy)

	// delete
	fmt.Println("delete TI", typeInstanceID)

	_, err = client.OnDelete(ctx, &pb.OnDeleteRequest{
		TypeinstanceId:       typeInstanceID,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	// last get

	res, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeinstanceId:       typeInstanceID,
		ResourceVersion:      1,
		AdditionalParameters: reqAdditionalParams,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("last get after delete - value", string(res.Value))
}
