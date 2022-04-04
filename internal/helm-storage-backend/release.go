package helmstoragebackend

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/release"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

var _ pb.ContextOnlyStorageBackendServer = &ReleaseHandler{}

const latestRevisionIndicator = 0

type (
	// ReleaseDetails holds Helm release details.
	ReleaseDetails struct {
		// Name specifies installed Helm release name.
		Name string `json:"name"`
		// Namespace specifies in which Kubernetes Namespace Helm release was installed.
		Namespace string `json:"namespace"`
		// Chart holds Helm Chart details.
		Chart ChartDetails `json:"chart"`
	}

	// ChartDetails holds Helm chart details.
	ChartDetails struct {
		// Name specifies Helm Chart name.
		Name string `json:"name"`
		// Version specifies the exact chart version.
		Version string `json:"version"`
		// Repo specifies URL where to locate the requested chart.
		Repo string `json:"repo"`
	}

	// ReleaseContext holds context used by Helm release storage backend.
	ReleaseContext struct {
		HelmRelease
		// ChartLocation specifies Helm Chart location.
		ChartLocation string `json:"chartLocation"`
	}
)

// ReleaseHandler handles incoming requests to the Helm release storage backend gRPC server.
type ReleaseHandler struct {
	pb.UnimplementedContextOnlyStorageBackendServer

	log     *zap.Logger
	fetcher *HelmReleaseFetcher
}

// NewReleaseHandler returns new ReleaseHandler.
func NewReleaseHandler(log *zap.Logger, helmRelFetcher *HelmReleaseFetcher) (*ReleaseHandler, error) {
	return &ReleaseHandler{
		log:     log,
		fetcher: helmRelFetcher,
	}, nil
}

// OnCreate checks whether a given Helm release is accessible this storage backend.
func (h *ReleaseHandler) OnCreate(_ context.Context, req *pb.OnCreateRequest) (*pb.OnCreateResponse, error) {
	if _, _, err := h.fetchHelmRelease(req.TypeInstanceId, req.Context); err != nil { // check if accessible
		return nil, err
	}
	return &pb.OnCreateResponse{}, nil
}

// GetValue returns a value for a given TypeInstance.
func (h *ReleaseHandler) GetValue(_ context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	rel, relCtx, err := h.fetchHelmRelease(req.TypeInstanceId, req.Context)
	if err != nil {
		return nil, err
	}

	releaseData := ReleaseDetails{
		Name:      rel.Name,
		Namespace: rel.Namespace,
		Chart: ChartDetails{
			Name:    rel.Chart.Metadata.Name,
			Version: rel.Chart.Metadata.Version,
			Repo:    relCtx.ChartLocation,
		},
	}

	value, err := json.Marshal(releaseData)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling response value")
	}

	return &pb.GetValueResponse{
		Value: value,
	}, nil
}

// OnUpdate checks whether a given Helm release is accessible this storage backend.
func (h *ReleaseHandler) OnUpdate(_ context.Context, req *pb.OnUpdateRequest) (*pb.OnUpdateResponse, error) {
	if _, _, err := h.fetchHelmRelease(req.TypeInstanceId, req.Context); err != nil { // check if accessible
		return nil, err
	}
	return &pb.OnUpdateResponse{}, nil
}

// OnDelete is NOP. Currently, we are not sure whether the release should be deleted or this should be more an information
// that someone wants to deregister this TypeInstance.
func (h *ReleaseHandler) OnDelete(_ context.Context, _ *pb.OnDeleteRequest) (*pb.OnDeleteResponse, error) {
	return &pb.OnDeleteResponse{}, nil
}

// GetLockedBy is NOP.
func (h *ReleaseHandler) GetLockedBy(_ context.Context, _ *pb.GetLockedByRequest) (*pb.GetLockedByResponse, error) {
	return &pb.GetLockedByResponse{}, nil
}

// OnLock is NOP.
func (h *ReleaseHandler) OnLock(_ context.Context, _ *pb.OnLockRequest) (*pb.OnLockResponse, error) {
	return &pb.OnLockResponse{}, nil
}

// OnUnlock is NOP.
func (h *ReleaseHandler) OnUnlock(_ context.Context, _ *pb.OnUnlockRequest) (*pb.OnUnlockResponse, error) {
	return &pb.OnUnlockResponse{}, nil
}

func (h *ReleaseHandler) getReleaseContext(contextBytes []byte) (*ReleaseContext, error) {
	var ctx ReleaseContext
	err := json.Unmarshal(contextBytes, &ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling context")
	}

	if ctx.Driver == nil {
		ctx.Driver = ptr.String(defaultHelmDriver)
	}

	return &ctx, nil
}

func (h *ReleaseHandler) fetchHelmRelease(ti string, ctx []byte) (*release.Release, *ReleaseContext, error) {
	relCtx, err := h.getReleaseContext(ctx)
	if err != nil {
		return nil, nil, gRPCInternalError(err)
	}

	rel, err := h.fetcher.FetchHelmRelease(ti, relCtx.HelmRelease)
	if err != nil {
		return nil, nil, err // it already handles grpc errors properly
	}

	return rel, relCtx, nil
}
