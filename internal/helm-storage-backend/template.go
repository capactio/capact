package helmstoragebackend

import (
	"context"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"go.uber.org/zap"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ pb.StorageBackendServer = &TemplateHandler{}

// TemplateHandler handles incoming requests to the Helm template storage backend gRPC server.
type TemplateHandler struct {
	pb.UnimplementedStorageBackendServer

	log *zap.Logger
}

// NewTemplateHandler returns new TemplateHandler.
func NewTemplateHandler(log *zap.Logger, helmCfgFlags *genericclioptions.ConfigFlags) *TemplateHandler {
	return &TemplateHandler{
		log: log,
	}
}

// GetValue returns a value for a given TypeInstance.
func (h *TemplateHandler) GetValue(_ context.Context, _ *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	h.log.Info("Getting value")
	return &pb.GetValueResponse{
		Value: []byte(`{"handler": "template"}`),
	}, nil
}
