package helmstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

var _ pb.StorageBackendServer = &ReleaseHandler{}

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
		// Name specifies Helm release name for a given request.
		Name string `json:"name"`
		// Namespace specifies in which Kubernetes Namespace Helm release is located.
		Namespace string `json:"namespace"`
		// ChartLocation specifies Helm Chart location.
		ChartLocation string `json:"chartLocation"`
		// Driver specifies drivers used for storing the Helm release.
		Driver *string `json:"driver,omitempty"`
	}
)

// ReleaseHandler handles incoming requests to the Helm release storage backend gRPC server.
type ReleaseHandler struct {
	pb.UnimplementedStorageBackendServer

	log                         *zap.Logger
	helmCfgFlags                *genericclioptions.ConfigFlags
	actionConfigurationProducer actionConfigurationProducerFn
}

// NewReleaseHandler returns new ReleaseHandler.
func NewReleaseHandler(log *zap.Logger, helmCfgFlags *genericclioptions.ConfigFlags) (*ReleaseHandler, error) {
	return &ReleaseHandler{
		log:                         log,
		helmCfgFlags:                helmCfgFlags,
		actionConfigurationProducer: ActionConfigurationProducer,
	}, nil
}

// OnCreate checks whether a given Helm release is accessible this storage backend.
func (h *ReleaseHandler) OnCreate(_ context.Context, req *pb.OnCreateRequest) (*pb.OnCreateResponse, error) {
	if err := h.checkIfHelmReleaseExist(req.TypeInstanceId, req.Context); err != nil {
		return nil, err
	}

	return &pb.OnCreateResponse{}, nil
}

// GetValue returns a value for a given TypeInstance.
func (h *ReleaseHandler) GetValue(_ context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	releaseContext, err := h.getReleaseContext(req.Context)
	if err != nil {
		return nil, err
	}

	helmGet, err := h.newHelmGet(h.helmCfgFlags, *releaseContext.Driver, releaseContext.Namespace)
	if err != nil {
		return nil, h.internalError(errors.Wrap(err, "while creating Helm get release client"))
	}

	// NOTE: req.resourceVersion is ignored on purpose.
	// Based on our contract we always return the latest Helm release revision.
	helmGet.Version = latestRevisionIndicator

	release, err := helmGet.Run(releaseContext.Name)
	switch {
	case err == nil:
	case errors.Is(err, driver.ErrReleaseNotFound):
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Helm release '%s/%s' for TypeInstance '%s' was not found", releaseContext.Namespace, releaseContext.Name, req.TypeInstanceId))
	default:
		return nil, h.internalError(errors.Wrap(err, "while getting Helm release"))
	}

	releaseData := ReleaseDetails{
		Name:      release.Name,
		Namespace: release.Namespace,
		Chart: ChartDetails{
			Name:    release.Chart.Metadata.Name,
			Version: release.Chart.Metadata.Version,
			Repo:    releaseContext.ChartLocation,
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
	if err := h.checkIfHelmReleaseExist(req.TypeInstanceId, req.Context); err != nil {
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
		return nil, h.internalError(errors.Wrap(err, "while unmarshaling context"))
	}

	if ctx.Driver == nil {
		ctx.Driver = ptr.String(defaultHelmDriver)
	}

	return &ctx, nil
}

func (h *ReleaseHandler) checkIfHelmReleaseExist(ti string, ctx []byte) error {
	releaseContext, err := h.getReleaseContext(ctx)
	if err != nil {
		return err
	}

	helmGet, err := h.newHelmGet(h.helmCfgFlags, *releaseContext.Driver, releaseContext.Namespace)
	if err != nil {
		return h.internalError(errors.Wrap(err, "while creating Helm get release client"))
	}

	_, err = helmGet.Run(releaseContext.Name)
	switch {
	case err == nil:
	case errors.Is(err, driver.ErrReleaseNotFound):
		return status.Error(codes.NotFound, fmt.Sprintf("Helm release '%s/%s' for TypeInstance '%s' was not found", releaseContext.Namespace, releaseContext.Name, ti))
	default:
		return h.internalError(errors.Wrap(err, "while checking if Helm release exists"))
	}

	return nil
}

func (h *ReleaseHandler) newHelmGet(flags *genericclioptions.ConfigFlags, driver, ns string) (*action.Get, error) {
	actionConfig, err := h.actionConfigurationProducer(flags, driver, ns)
	if err != nil {
		return nil, err
	}

	return action.NewGet(actionConfig), nil
}

func (h *ReleaseHandler) internalError(err error) error {
	return status.Error(codes.Internal, err.Error())
}
