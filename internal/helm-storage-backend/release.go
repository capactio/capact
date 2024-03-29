package helmstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/release"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

var _ pb.ContextStorageBackendServer = &ReleaseHandler{}

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
	pb.UnimplementedContextStorageBackendServer

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
	if _, _, err := h.fetchHelmReleaseForTI(req.TypeInstanceId, req.Context); err != nil { // check if accessible
		return nil, err
	}
	return &pb.OnCreateResponse{}, nil
}

// GetPreCreateValue returns a value for a given context without TypeInstance details.
func (h *ReleaseHandler) GetPreCreateValue(ctx context.Context, req *pb.GetPreCreateValueRequest) (*pb.GetPreCreateValueResponse, error) {
	rel, relCtx, err := h.fetchHelmReleaseForPrecreate(req.Context)
	if err != nil {
		return nil, err
	}

	value, err := h.marshalReleaseDetails(rel, relCtx)
	if err != nil {
		return nil, err
	}

	return &pb.GetPreCreateValueResponse{
		Value: value,
	}, nil
}

// GetValue returns a value for a given TypeInstance.
func (h *ReleaseHandler) GetValue(_ context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	rel, relCtx, err := h.fetchHelmReleaseForTI(req.TypeInstanceId, req.Context)
	if err != nil {
		return nil, err
	}

	value, err := h.marshalReleaseDetails(rel, relCtx)
	if err != nil {
		return nil, err
	}

	return &pb.GetValueResponse{
		Value: value,
	}, nil
}

// OnUpdate checks whether a given Helm release is accessible this storage backend.
func (h *ReleaseHandler) OnUpdate(_ context.Context, req *pb.OnUpdateRequest) (*pb.OnUpdateResponse, error) {
	if _, _, err := h.fetchHelmReleaseForTI(req.TypeInstanceId, req.Context); err != nil { // check if accessible
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

// OnDeleteRevision is NOP.
func (*ReleaseHandler) OnDeleteRevision(context.Context, *pb.OnDeleteRevisionRequest) (*pb.OnDeleteRevisionResponse, error) {
	return &pb.OnDeleteRevisionResponse{}, nil
}

func (h *ReleaseHandler) marshalReleaseDetails(rel *release.Release, relCtx *ReleaseContext) ([]byte, error) {
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
		return nil, gRPCInternalError(errors.Wrap(err, "while marshaling response value"))
	}

	return value, nil
}

func (h *ReleaseHandler) fetchHelmReleaseForPrecreate(relCtx []byte) (*release.Release, *ReleaseContext, error) {
	additionalCtxMsg := "TypeInstance ID: not yet known"
	return h.fetchHelmRelease(relCtx, additionalCtxMsg)
}

func (h *ReleaseHandler) fetchHelmReleaseForTI(ti string, relCtx []byte) (*release.Release, *ReleaseContext, error) {
	additionalCtxMsg := fmt.Sprintf("TypeInstance ID: '%s'", ti)
	return h.fetchHelmRelease(relCtx, additionalCtxMsg)
}

func (h *ReleaseHandler) fetchHelmRelease(relCtx []byte, additionalCtxMsg string) (*release.Release, *ReleaseContext, error) {
	resolvedRelCtx, err := h.getReleaseContext(relCtx)
	if err != nil {
		return nil, nil, gRPCInternalError(err)
	}

	rel, err := h.fetcher.FetchHelmRelease(resolvedRelCtx.HelmRelease, ptr.String(additionalCtxMsg))
	if err != nil {
		return nil, nil, err // it already handles grpc errors properly
	}

	return rel, resolvedRelCtx, nil
}
