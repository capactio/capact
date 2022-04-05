package argoactions

import (
	"context"
	"encoding/json"

	storagebackend "capact.io/capact/pkg/hub/storage-backend"

	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/local"
	"github.com/pkg/errors"
)

func resolveBackendsValues(ctx context.Context, client *hubclient.Client, ids []string) (map[string]storagebackend.TypeValue, error) {
	// get values
	tiMap, err := client.FindTypeInstances(ctx, ids, local.WithFields(local.TypeInstanceRootFields|local.TypeInstanceLatestResourceVersionValueField))
	if err != nil {
		return nil, errors.Wrap(err, "while fetching TypeInstance values")
	}

	// create result
	result := make(map[string]storagebackend.TypeValue)
	for id, ti := range tiMap {
		if ti.LatestResourceVersion == nil || ti.LatestResourceVersion.Spec == nil {
			continue
		}

		data, err := json.Marshal(ti.LatestResourceVersion.Spec.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "while marshaling TypeInstance value for ID %q", id)
		}

		var value storagebackend.TypeValue
		err = json.Unmarshal(data, &value)
		if err != nil {
			return nil, errors.Wrapf(err, "while unmarshaling TypeInstance value for ID %q", id)
		}

		result[id] = value
	}

	return result, nil
}
