package helmstoragebackend

import (
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
