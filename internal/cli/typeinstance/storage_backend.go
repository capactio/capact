package typeinstance

import (
	"context"
	"encoding/json"

	"capact.io/capact/internal/cli/client"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"github.com/pkg/errors"
)

// StorageBackendData holds the information about storage backend data.
type StorageBackendData struct {
	URL           string      `json:"url"`
	AcceptValue   bool        `json:"acceptValue"`
	ContextSchema interface{} `json:"contextSchema"`
}

// NewStorageBackendData returns a new StorageBackendData instance based on backend used by passed TypeInstance.
func NewStorageBackendData(ctx context.Context, cli client.Hub, typeInstance *gqllocalapi.TypeInstance) (*StorageBackendData, error) {
	var backendValue *StorageBackendData
	if typeInstance.Backend == nil || typeInstance.Backend.Abstract {
		return nil, nil
	}
	backendTI, err := cli.FindTypeInstance(ctx, typeInstance.Backend.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while finding backend TypeInstance")
	}
	valueBytes, err := json.Marshal(backendTI.LatestResourceVersion.Spec.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling storage backend value")
	}
	err = json.Unmarshal(valueBytes, &backendValue)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling storage backend value")
	}
	return backendValue, nil
}
