package secretstoragebackend

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	tellercore "github.com/spectralops/teller/pkg/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

// Context holds Secret storage backend specific parameters.
type Context struct {
	Provider string `json:"provider"`
}

// Providers holds map of the configured providers for the Handler.
type Providers map[string]tellercore.Provider

// Count returns the providers count.
func (p Providers) Count() int {
	return len(p)
}

// GetDefault returns the default provider in case there is exactly one configured.
func (p Providers) GetDefault() (tellercore.Provider, error) {
	count := p.Count()
	invalidCountErr := fmt.Errorf("invalid number of providers configured to get default one: expected: 1, actual: %d", count)

	if count > 1 {
		return nil, invalidCountErr
	}

	for _, provider := range p {
		// return first one
		return provider, nil
	}

	// empty map - no providers
	return nil, invalidCountErr
}

var _ pb.ValueAndContextStorageBackendServer = &Handler{}

// Handler handles incoming requests to the Secret storage backend gRPC server.
type Handler struct {
	pb.UnimplementedValueAndContextStorageBackendServer

	log *zap.Logger

	providers Providers
}

const (
	lockedByField        = "locked_by"
	firstResourceVersion = 1
)

var (
	// NilRequestInputError describes an error with an invalid request.
	NilRequestInputError = status.Error(codes.InvalidArgument, "request data cannot be nil")
)

// NewHandler returns new Handler.
func NewHandler(log *zap.Logger, providers Providers) *Handler {
	return &Handler{
		log:       log,
		providers: providers,
	}
}

// GetValue returns a value for a given TypeInstance. It returns nil as value if a given secret is not found.
func (h *Handler) GetValue(_ context.Context, request *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForTypeInstanceValue(provider, request.TypeInstanceId, request.ResourceVersion)
	entry, err := h.getEntry(provider, key)
	if err != nil {
		return nil, err
	}

	if !entry.IsFound {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("TypeInstance %q in revision %d was not found", request.TypeInstanceId, request.ResourceVersion))
	}

	return &pb.GetValueResponse{
		Value: []byte(entry.Value),
	}, nil
}

// GetLockedBy returns a locked by data for a given TypeInstance. It returns nil as value if a given secret is not found.
func (h *Handler) GetLockedBy(_ context.Context, request *pb.GetLockedByRequest) (*pb.GetLockedByResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := tellercore.KeyPath{
		Path: h.storagePathForTypeInstance(provider, request.TypeInstanceId),
	}
	entries, err := h.getEntriesForPath(provider, key)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("TypeInstance %q not found: secret from path %q is empty", request.TypeInstanceId, key.Path))
	}

	var lockedBy *string
	for _, entry := range entries {
		if entry.Key == lockedByField && entry.Value != "" {
			lockedBy = ptr.String(entry.Value)
		}
	}

	return &pb.GetLockedByResponse{
		LockedBy: lockedBy,
	}, nil
}

// OnCreate handles TypeInstance creation by creating secret in a given provider.
func (h *Handler) OnCreate(_ context.Context, request *pb.OnCreateValueAndContextRequest) (*pb.OnCreateResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForTypeInstanceValue(provider, request.TypeInstanceId, firstResourceVersion)
	if err := h.ensureSecretCanBeCreated(provider, key); err != nil {
		return nil, err
	}

	err = h.putEntry(provider, key, request.Value)
	if err != nil {
		return nil, err
	}

	return &pb.OnCreateResponse{}, nil
}

// OnUpdate handles TypeInstance update by updating secret in a given provider.
func (h *Handler) OnUpdate(_ context.Context, request *pb.OnUpdateValueAndContextRequest) (*pb.OnUpdateResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForTypeInstanceValue(provider, request.TypeInstanceId, request.NewResourceVersion)

	if err := h.ensureSecretCanBeUpdated(provider, key, request.OwnerId); err != nil {
		return nil, err
	}

	err = h.putEntry(provider, key, request.NewValue)
	if err != nil {
		return nil, err
	}

	return &pb.OnUpdateResponse{}, nil
}

// OnLock handles TypeInstance locking by setting a secret entry in a given provider.
// It doesn't check whether a given TypeInstance is already locked, but overrides the value in place
func (h *Handler) OnLock(_ context.Context, request *pb.OnLockRequest) (*pb.OnLockResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	err = h.ensureSecretIsNotLocked(provider, request.TypeInstanceId)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForLockedBy(provider, request.TypeInstanceId)

	err = h.putEntry(provider, key, []byte(request.LockedBy))
	if err != nil {
		return nil, err
	}

	return &pb.OnLockResponse{}, nil
}

// OnUnlock handles TypeInstance unlocking by removing secret entry in a given provider.
func (h *Handler) OnUnlock(_ context.Context, request *pb.OnUnlockRequest) (*pb.OnUnlockResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := tellercore.KeyPath{
		Path: h.storagePathForTypeInstance(provider, request.TypeInstanceId),
	}
	err = h.ensureSecretCanBeUnlocked(provider, key)
	if err != nil {
		return nil, err
	}

	lockedByKey := h.storageKeyForLockedBy(provider, request.TypeInstanceId)
	err = h.deleteEntry(provider, lockedByKey)
	if err != nil {
		return nil, err
	}

	return &pb.OnUnlockResponse{}, nil
}

// OnDelete handles TypeInstance deletion by removing a secret in a given provider.
// It checks whether a given TypeInstance is locked before doing such operation.
func (h *Handler) OnDelete(_ context.Context, request *pb.OnDeleteValueAndContextRequest) (*pb.OnDeleteResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := tellercore.KeyPath{
		Path: h.storagePathForTypeInstance(provider, request.TypeInstanceId),
	}
	err = h.ensureSecretCanBeDeleted(provider, key, request.OwnerId)
	if err != nil {
		return nil, err
	}

	err = provider.DeleteMapping(key)
	if err != nil {
		return nil, h.internalError(errors.Wrapf(err, "while deleting TypeInstance %q", request.TypeInstanceId))
	}

	return &pb.OnDeleteResponse{}, nil
}

// OnDeleteRevision handles TypeInstance's revision deletion by removing a secret entry in a given provider.
// It checks whether a given TypeInstance is locked before doing such operation.
func (h *Handler) OnDeleteRevision(ctx context.Context, request *pb.OnDeleteRevisionRequest) (*pb.OnDeleteRevisionResponse, error) {
	if request == nil {
		return nil, NilRequestInputError
	}

	provider, err := h.getProviderFromContext(request.Context)
	if err != nil {
		return nil, err
	}

	key := h.storageKeyForTypeInstanceValue(provider, request.TypeInstanceId, request.ResourceVersion)

	entry, err := h.getEntry(provider, key)
	if err != nil {
		return nil, err
	}
	if !entry.IsFound {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("TypeInstance %q in revision %d was not found", request.TypeInstanceId, request.ResourceVersion))
	}

	if err := h.ensureSecretCanBeDeleted(provider, key, request.OwnerId); err != nil {
		return nil, err
	}

	err = h.deleteEntry(provider, key)
	if err != nil {
		return nil, err
	}

	return &pb.OnDeleteRevisionResponse{}, nil
}

func (h *Handler) getProviderFromContext(contextBytes []byte) (tellercore.Provider, error) {
	if len(contextBytes) == 0 {
		provider, err := h.providers.GetDefault()
		if err != nil {
			return nil, h.failedPreconditionError(errors.Wrap(err, "while getting default provider based on empty context"))
		}
		return provider, nil
	}

	var context Context
	err := json.Unmarshal(contextBytes, &context)
	if err != nil {
		return nil, h.internalError(errors.Wrap(err, "while unmarshaling context"))
	}

	if context.Provider == "" {
		provider, err := h.providers.GetDefault()
		if err != nil {
			return nil, h.failedPreconditionError(errors.Wrap(err, "while getting default provider as not specified in context"))
		}

		return provider, nil
	}

	provider, ok := h.providers[context.Provider]
	if !ok {
		return nil, h.failedPreconditionError(fmt.Errorf("missing loaded provider with name %q", context.Provider))
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

func (h *Handler) getEntriesForPath(provider tellercore.Provider, key tellercore.KeyPath) ([]tellercore.EnvEntry, error) {
	h.log.Info("getting whole secret", zap.String("path", key.Path), zap.String("provider", provider.Name()))
	entries, err := provider.GetMapping(key)
	if err != nil {
		return nil, h.internalError(errors.Wrapf(err, "while getting value by path %q", key.Path))
	}

	return entries, nil
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
	default:
		prefix = "/"
	}

	return fmt.Sprintf("%scapact/%s", prefix, tiID)
}

func (h *Handler) ensureSecretCanBeCreated(provider tellercore.Provider, key tellercore.KeyPath) error {
	entries, err := h.getEntriesForPath(provider, key)
	if err != nil {
		return h.internalError(err)
	}

	if len(entries) != 0 {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("path %q in provider %q already exist", key.Path, provider.Name()))
	}

	return nil
}

func (h *Handler) ensureSecretCanBeUpdated(provider tellercore.Provider, key tellercore.KeyPath, ownerID *string) error {
	entries, err := h.getEntriesForPath(provider, key)
	if err != nil {
		return h.internalError(err)
	}

	if len(entries) == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("path %q in provider %q not found", key.Path, provider.Name()))
	}

	for _, entry := range entries {
		if entry.Key == lockedByField {
			if ownerID != nil && entry.Value == *ownerID {
				continue
			}
			return h.typeInstanceLockedError(key.Path, entry.Value)
		}
		if entry.Key == key.Field {
			return status.Error(codes.AlreadyExists, fmt.Sprintf("field %q for path %q in provider %q already exist", key.Field, key.Path, provider.Name()))
		}
	}

	return nil
}

func (h *Handler) ensureSecretIsNotLocked(provider tellercore.Provider, typeInstanceID string) error {
	key := h.storageKeyForLockedBy(provider, typeInstanceID)
	entry, err := h.getEntry(provider, key)
	if err != nil {
		return h.internalError(errors.Wrapf(err, "while getting entry"))
	}
	if entry.IsFound && entry.Value != "" {
		return h.typeInstanceLockedError(key.Path, entry.Value)
	}

	return nil
}

func (h *Handler) ensureSecretCanBeDeleted(provider tellercore.Provider, key tellercore.KeyPath, ownerID *string) error {
	entries, err := h.getEntriesForPath(provider, key)
	if err != nil {
		return h.internalError(err)
	}

	if len(entries) == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("path %q in provider %q not found", key.Path, provider.Name()))
	}

	for _, entry := range entries {
		if entry.Key != lockedByField {
			continue
		}
		if ownerID != nil && entry.Value == *ownerID {
			continue
		}
		return h.typeInstanceLockedError(key.Path, entry.Value)
	}

	return nil
}

func (h *Handler) ensureSecretCanBeUnlocked(provider tellercore.Provider, key tellercore.KeyPath) error {
	entries, err := h.getEntriesForPath(provider, key)
	if err != nil {
		return h.internalError(err)
	}

	if len(entries) == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("path %q in provider %q not found", key.Path, provider.Name()))
	}

	return nil
}

func (h *Handler) internalError(err error) error {
	return status.Error(codes.Internal, err.Error())
}

func (h *Handler) failedPreconditionError(err error) error {
	return status.Error(codes.FailedPrecondition, err.Error())
}

func (h *Handler) typeInstanceLockedError(path, lockedByValue string) error {
	return h.failedPreconditionError(fmt.Errorf("typeInstance locked: path %q contains %q property with value %q", path, lockedByField, lockedByValue))
}
