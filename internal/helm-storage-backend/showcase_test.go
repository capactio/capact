package helmstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

// To run this test, execute:
// GRPC_SECRET_STORAGE_BACKEND_ADDR=":50051" go test ./internal/helm-storage-backend/... -run "^TestShowcase$" -v -count 1
func TestShowcase(t *testing.T) {
	srvAddr := os.Getenv("GRPC_SECRET_STORAGE_BACKEND_ADDR")
	if srvAddr == "" {
		t.Skip()
	}

	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	require.NoError(t, err)

	ctx := context.Background()
	client := pb.NewStorageBackendClient(conn)

	// ===== GET =====
	out, err := client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId: "42",
		Context: mustMarshal(t, ReleaseContext{
			Name:          "example-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})

	require.NoError(t, err)

	details := &ReleaseDetails{}
	require.NoError(t, json.Unmarshal(out.Value, details))

	fmt.Printf("GetValue for valid release")
	fmt.Printf("\t\t Name: %s\n", details.Name)
	fmt.Printf("\t\t Namespace: %s\n", details.Namespace)
	fmt.Printf("\t\t Chart.Name: %s\n", details.Chart.Name)
	fmt.Printf("\t\t Chart.Version: %s\n", details.Chart.Version)
	fmt.Printf("\t\t Chart.Repo: %s\n", details.Chart.Repo)

	_, err = client.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId: "42",
		Context: mustMarshal(t, ReleaseContext{
			Name:          "fake-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})

	fmt.Printf("GetValue err if release doesn't exist: %v\n", err)

	// ===== UPDATE =====
	_, err = client.OnUpdate(ctx, &pb.OnUpdateRequest{
		Context: mustMarshal(t, ReleaseContext{
			Name:          "example-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})
	fmt.Printf("OnUpdate err if release exists: %v\n", err)

	_, err = client.OnUpdate(ctx, &pb.OnUpdateRequest{
		TypeInstanceId: "42",
		Context: mustMarshal(t, ReleaseContext{
			Name:          "fake-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})
	fmt.Printf("OnUpdate err if release doesn't exist: %v\n", err)

	// ===== CREATE =====
	_, err = client.OnCreate(ctx, &pb.OnCreateRequest{
		Context: mustMarshal(t, ReleaseContext{
			Name:          "example-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})
	fmt.Printf("OnCreate err if release exists: %v\n", err)

	_, err = client.OnCreate(ctx, &pb.OnCreateRequest{
		TypeInstanceId: "42",
		Context: mustMarshal(t, ReleaseContext{
			Name:          "fake-release",
			Namespace:     "default",
			ChartLocation: "https://charts.bitnami.com/bitnami",
		})})
	fmt.Printf("OnCreate err if release doesn't exist: %v\n", err)
}
