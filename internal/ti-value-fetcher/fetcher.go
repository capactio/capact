package tivaluefetcher

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"sigs.k8s.io/yaml"

	storagebackend "capact.io/capact/pkg/hub/storage-backend"

	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TypeInstanceData represents the TypeInstance data in workflow.
type TypeInstanceData struct {
	Value   *json.RawMessage `json:"value"`
	Backend *Backend         `json:"backend"`
}

// Backend represents the TypeInstance Backend.
type Backend struct {
	Context json.RawMessage `json:"context"`
}

// DefaultFilePermissions are the default file permissions
// of the output artifact files created by the runners.
const DefaultFilePermissions = 0644

// TIValueFetcher resolves TypeInstance value before it is being actually created.
type TIValueFetcher struct {
	log *zap.Logger
}

// New creates new TIValueFetcher.
func New(log *zap.Logger) *TIValueFetcher {
	return &TIValueFetcher{log: log}
}

// Do resolves TypeInstance value based on the artifact and backend data. It calls external storage backend.
func (r *TIValueFetcher) Do(ctx context.Context, typeInstanceData TypeInstanceData, storageBackendValue storagebackend.TypeInstanceValue, dialOpts ...grpc.DialOption) (TypeInstanceData, error) {
	if typeInstanceData.Value != nil {
		// nothing to do here - save input as output
		r.log.Info("Value already provided. Skipping...")
		return typeInstanceData, nil
	}

	if storageBackendValue.AcceptValue {
		r.log.Info("Storage backend accepts value. Skipping...")
		return typeInstanceData, nil
	}

	// create client
	conn, err := grpc.Dial(storageBackendValue.URL, dialOpts...)
	if err != nil {
		return TypeInstanceData{}, errors.Wrapf(err, "while setting up connection to storage backend %q", storageBackendValue.URL)
	}
	client := pb.NewContextStorageBackendClient(conn)

	// do call
	res, err := client.GetPreCreateValue(ctx, &pb.GetPreCreateValueRequest{Context: typeInstanceData.Backend.Context})
	if err != nil {
		return TypeInstanceData{}, errors.Wrapf(err, "while getting precreate value from storage backend %q", storageBackendValue.URL)
	}

	newValue := json.RawMessage(res.Value)
	typeInstanceData.Value = &newValue

	return typeInstanceData, nil
}

// LoadFromFile loads input for TIValueFetcher.Do from files.
func (r *TIValueFetcher) LoadFromFile(tiArtifactFilePath, storageBackendFilePath string) (TypeInstanceData, storagebackend.TypeInstanceValue, error) {
	var tiArtifact TypeInstanceData
	err := r.unmarshalFromFile(tiArtifactFilePath, &tiArtifact)
	if err != nil {
		return TypeInstanceData{}, storagebackend.TypeInstanceValue{}, err
	}

	var storageBackend struct {
		Value storagebackend.TypeInstanceValue `json:"value"`
	}
	err = r.unmarshalFromFile(storageBackendFilePath, &storageBackend)
	if err != nil {
		return TypeInstanceData{}, storagebackend.TypeInstanceValue{}, err
	}

	return tiArtifact, storageBackend.Value, nil
}

func (r *TIValueFetcher) unmarshalFromFile(path string, out interface{}) error {
	bytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return errors.Wrapf(err, "while reading file from path %q", path)
	}

	if err := yaml.Unmarshal(bytes, &out); err != nil {
		return errors.Wrapf(err, "while unmarshaling data from file %q", path)
	}

	return nil
}

// SaveToFile saves the output to a file under the path.
func (r *TIValueFetcher) SaveToFile(path string, tiData TypeInstanceData) error {
	bytes, err := yaml.Marshal(tiData)
	if err != nil {
		return errors.Wrap(err, "while marshaling output to bytes")
	}

	err = ioutil.WriteFile(path, bytes, DefaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to %q", path)
	}
	return nil
}
