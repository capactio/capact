package helmstoragebackend

import (
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
