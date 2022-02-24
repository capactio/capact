package secretstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"github.com/pkg/errors"
	tellercore "github.com/spectralops/teller/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdditionalParameters struct {
	Provider string `json:"provider"`
}

var _ pb.StorageBackendServer = &Handler{}

type Handler struct {
	pb.UnimplementedStorageBackendServer

	log *zap.Logger

	providers map[string]tellercore.Provider
}

const (
	lockedByField        = "locked_by"
	firstResourceVersion = 1
)

var (
	NilRequestInputError = status.Error(codes.InvalidArgument, "request data cannot be nil")
)

func NewHandler(log *zap.Logger, providers map[string]tellercore.Provider) *Handler {
	return &Handler{
		log:       log,
		providers: providers,
	}
}

func (h *Handler) GetValue(_ context.Context, request *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromAdditionalParams(request.AdditionalParameters)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForTypeInstanceValue(provider, request.TypeinstanceId, request.ResourceVersion)
	entry, err := h.getEntry(provider, key)
	if err != nil {
		return nil, err
	}

	var value []byte
	if entry.IsFound {
		value = []byte(entry.Value)
	}

	return &pb.GetValueResponse{
		Value: value,
	}, nil
}

func (h *Handler) GetLockedBy(_ context.Context, request *pb.GetLockedByRequest) (*pb.GetLockedByResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromAdditionalParams(request.AdditionalParameters)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForLockedBy(provider, request.TypeinstanceId)
	entry, err := h.getEntry(provider, key)
	if err != nil {
		return nil, err
	}

	var lockedBy *string
	if entry.IsFound && entry.Value != "" {
		lockedBy = ptr.String(entry.Value)
	}

	return &pb.GetLockedByResponse{
		LockedBy: lockedBy,
	}, nil
}

func (h *Handler) OnCreate(_ context.Context, request *pb.OnCreateRequest) (*pb.OnCreateResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	err := h.handlePutValue(
		request.AdditionalParameters,
		request.TypeinstanceId,
		firstResourceVersion,
		request.Value,
	)
	if err != nil {
		return nil, err
	}

	return &pb.OnCreateResponse{}, nil
}

func (h *Handler) OnUpdate(_ context.Context, request *pb.OnUpdateRequest) (*pb.OnUpdateResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	err := h.handlePutValue(
		request.AdditionalParameters,
		request.TypeinstanceId,
		request.NewResourceVersion,
		request.NewValue,
	)
	if err != nil {
		return nil, err
	}

	return &pb.OnUpdateResponse{}, nil
}

// OnLock doesn't check whether a given TypeInstance is already locked, but overrides the value in place
// TODO(review): Is that valid assumption? Is there a need to complicate the flow here?
func (h *Handler) OnLock(_ context.Context, request *pb.OnLockRequest) (*pb.OnLockResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromAdditionalParams(request.AdditionalParameters)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForLockedBy(provider, request.TypeinstanceId)
	err = h.putEntry(provider, key, []byte(request.LockedBy))
	if err != nil {
		return nil, err
	}

	return &pb.OnLockResponse{}, nil
}

func (h *Handler) OnUnlock(_ context.Context, request *pb.OnUnlockRequest) (*pb.OnUnlockResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromAdditionalParams(request.AdditionalParameters)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForLockedBy(provider, request.TypeinstanceId)
	err = h.deleteEntry(provider, key)
	if err != nil {
		return nil, err
	}

	return &pb.OnUnlockResponse{}, nil
}

// OnDelete doesn't check whether a given TypeInstance is locked. It assumes the caller ensured it's unlocked state.
// TODO(review): Is that a valid assumption?
func (h *Handler) OnDelete(_ context.Context, request *pb.OnDeleteRequest) (*pb.OnDeleteResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromAdditionalParams(request.AdditionalParameters)
	if err != nil {
		return nil, err
	}

	err = provider.DeleteMapping(tellercore.KeyPath{
		Path: h.storagePathForTypeInstance(provider, request.TypeinstanceId),
	})
	if err != nil {
		return nil, h.internalError(errors.Wrapf(err, "while deleting TypeInstance %q", request.TypeinstanceId))
	}

	return &pb.OnDeleteResponse{}, nil
}

func (h *Handler) handlePutValue(additionalParamsBytes []byte, typeInstanceID string, resourceVersion uint32, value []byte) error {
	provider, err := h.getProviderFromAdditionalParams(additionalParamsBytes)
	if err != nil {
		return err
	}

	key := h.storageKeyForTypeInstanceValue(provider, typeInstanceID, resourceVersion)

	if err := h.ensureEntryDoesNotExist(provider, key); err != nil {
		return err
	}

	err = h.putEntry(provider, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) getProviderFromAdditionalParams(additionalParamsBytes []byte) (tellercore.Provider, error) {
	var additionalParams AdditionalParameters
	err := json.Unmarshal(additionalParamsBytes, &additionalParams)
	if err != nil {
		return nil, h.internalError(errors.Wrap(err, "while unmarshaling additional parameters"))
	}

	provider, ok := h.providers[additionalParams.Provider]
	if !ok {
		return nil, h.internalError(fmt.Errorf("missing loaded provider with name %q", additionalParams.Provider))
	}

	return provider, nil
}

func (h *Handler) getEntry(provider tellercore.Provider, key tellercore.KeyPath) (*tellercore.EnvEntry, error) {
	h.log.Info("getting entry", zap.String("path", key.Path), zap.String("provider", provider.Name()))
	entry, err := provider.Get(key)
	if err != nil {
		return nil, h.internalError(errors.Wrapf(err, "while getting value by key %q", key.Path))
	}

	return entry, nil
}

func (h *Handler) putEntry(provider tellercore.Provider, key tellercore.KeyPath, value []byte) error {
	h.log.Info("putting entry", zap.String("path", key.Path), zap.String("field", key.Field), zap.String("provider", provider.Name()))
	err := provider.Put(key, string(value))
	if err != nil {
		return h.internalError(errors.Wrapf(err, "while putting value for key %q", key.Path))
	}

	return nil
}

func (h *Handler) deleteEntry(provider tellercore.Provider, key tellercore.KeyPath) error {
	h.log.Info("deleting entry", zap.String("path", key.Path), zap.String("field", key.Field), zap.String("provider", provider.Name()))
	err := provider.Delete(key)
	if err != nil {
		return h.internalError(errors.Wrapf(err, "while deleting %q for key %q", key.Field, key.Path))
	}

	return nil
}

func (h *Handler) storageKeyForTypeInstanceValue(provider tellercore.Provider, tiID string, tiResourceVersion uint32) tellercore.KeyPath {
	return tellercore.KeyPath{
		Path:  h.storagePathForTypeInstance(provider, tiID),
		Field: strconv.Itoa(int(tiResourceVersion)),
	}
}

func (h *Handler) storageKeyForLockedBy(provider tellercore.Provider, tiID string) tellercore.KeyPath {
	return tellercore.KeyPath{
		Path:  h.storagePathForTypeInstance(provider, tiID),
		Field: lockedByField,
	}
}

func (h *Handler) storagePathForTypeInstance(provider tellercore.Provider, tiID string) string {
	// depending on provider there might be a different path format
	// e.g. see https://github.com/SpectralOps/teller#google-secret-manager

	var prefix string
	switch provider.Name() {
	case "dotenv":
		prefix = "/tmp/"
	}

	return fmt.Sprintf("%s/capact/%s", prefix, tiID)
}

func (h *Handler) ensureEntryDoesNotExist(provider tellercore.Provider, key tellercore.KeyPath) error {
	entry, err := h.getEntry(provider, key)
	if err != nil {
		return h.internalError(errors.Wrapf(err, "while getting entry for path %q", key.Path))
	}
	if entry.IsFound {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("entry %q in provider %q already exist", key.Path, provider.Name()))
	}

	return nil
}

func (h *Handler) internalError(err error) error {
	return status.Error(codes.Internal, err.Error())
}
