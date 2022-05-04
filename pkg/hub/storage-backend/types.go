package storagebackend

import (
	"context"
	"encoding/json"

	"capact.io/capact/internal/cli/client"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"github.com/pkg/errors"
)

// TypeInstanceValue defines properties for TypeInstance value for every Storage Backend.
type TypeInstanceValue struct {
	URL           string      `json:"url"`
	AcceptValue   bool        `json:"acceptValue"`
	ContextSchema interface{} `json:"contextSchema"`
}

// NewTypeInstanceValue returns a new TypeInstanceValue instance based on backend used by passed TypeInstance.
func NewTypeInstanceValue(ctx context.Context, cli client.Hub, typeInstance *gqllocalapi.TypeInstance) (*TypeInstanceValue, error) {
	var typeInstanceValue *TypeInstanceValue
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
	err = json.Unmarshal(valueBytes, &typeInstanceValue)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling storage backend value")
	}
	return typeInstanceValue, nil
}
