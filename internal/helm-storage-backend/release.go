package helmstoragebackend

import (
	"context"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"go.uber.org/zap"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ pb.StorageBackendServer = &ReleaseHandler{}

// ReleaseHandler handles incoming requests to the Helm release storage backend gRPC server.
type ReleaseHandler struct {
	pb.UnimplementedStorageBackendServer

	log *zap.Logger
}

// NewReleaseHandler returns new ReleaseHandler.
func NewReleaseHandler(log *zap.Logger, helmCfgFlags *genericclioptions.ConfigFlags) *ReleaseHandler {
	return &ReleaseHandler{
		log: log,
	}
}

// GetValue returns a value for a given TypeInstance.
func (h *ReleaseHandler) GetValue(_ context.Context, _ *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	h.log.Info("Getting value")
	return &pb.GetValueResponse{
		Value: []byte(`{"handler": "release"}`),
	}, nil
}
