package argoactions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	storage_backend "capact.io/capact/pkg/hub/storage-backend"

	"capact.io/capact/pkg/hub/client/local"

	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/sdk/validation"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

// UpdateAction represents the update TypeInstancess action.
const UpdateAction = "UpdateAction"

// UpdateConfig stores the configuration parameters for update TypeInstances action.
type UpdateConfig struct {
	PayloadFilepath  string
	TypeInstancesDir string
}

// Update implements the Action interface.
// It is used to update existing TypeInstances in the Local Hub.
type Update struct {
	log    *zap.Logger
	client *hubclient.Client
	cfg    UpdateConfig
}

// NewUpdateAction returns a new Update instance.
func NewUpdateAction(log *zap.Logger, client *hubclient.Client, cfg UpdateConfig) Action {
	return &Update{
		log:    log,
		client: client,
		cfg:    cfg,
	}
}

// Do updates existing TypeInstances in the Local Hub.
func (u *Update) Do(ctx context.Context) error {
	payloadBytes, err := ioutil.ReadFile(u.cfg.PayloadFilepath)
	if err != nil {
		return errors.Wrap(err, "while reading payload file")
	}

	var payload []graphqllocal.UpdateTypeInstancesInput
	if err := yaml.Unmarshal(payloadBytes, &payload); err != nil {
		return errors.Wrap(err, "while unmarshaling payload bytes")
	}

	if len(payload) == 0 {
		u.log.Info("No TypeInstances to update")
		return nil
	}

	files, err := ioutil.ReadDir(u.cfg.TypeInstancesDir)
	if err != nil {
		return errors.Wrap(err, "while listing Type Instances directory")
	}

	typeInstanceValues := map[string]map[string]interface{}{}

	for _, f := range files {
		path := path.Join(u.cfg.TypeInstancesDir, f.Name())

		typeInstanceValueBytes, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return errors.Wrapf(err, "while reading TypeInstance value file %s", path)
		}

		values := map[string]interface{}{}
		if err := yaml.Unmarshal(typeInstanceValueBytes, &values); err != nil {
			return errors.Wrapf(err, "while unmarshaling bytes from %s file", path)
		}

		typeInstanceValues[f.Name()] = values
	}

	backends, mapping, err := u.resolveBackendsValues(ctx, payload)
	if err != nil {
		return errors.Wrap(err, "while resolving storage backends values")
	}

	u.log.Info("Rendering update TypeInstance input")
	if err := u.render(payload, typeInstanceValues, u.shouldIncludeTIValueFn(backends, mapping)); err != nil {
		return errors.Wrap(err, "while rendering UpdateTypeInstancesInput")
	}

	u.log.Info("Validating TypeInstances")

	r := validation.ResultAggregator{}
	err = r.Report(validation.ValidateTypeInstanceToUpdate(ctx, u.client, payload))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstance")
	}
	if r.ErrorOrNil() != nil {
		return r.ErrorOrNil()
	}

	u.log.Info("Updating TypeInstances in Hub...", zap.Int("TypeInstance count", len(payload)))

	uploadOutput, err := u.updateTypeInstances(ctx, payload)
	if err != nil {
		return errors.Wrap(err, "while updating TypeInstances")
	}

	for _, ti := range uploadOutput {
		u.log.Info("TypeInstance updated", zap.String("ID", ti.ID))
	}

	return nil
}

type (
	tiToUpdateID string
	backendTIID  string
)

func (u *Update) resolveBackendsValues(ctx context.Context, typeInstances []graphqllocal.UpdateTypeInstancesInput) (map[string]storage_backend.TypeValue, map[tiToUpdateID]backendTIID, error) {
	// get IDs of TypeInstances to update
	var tiToUpdateIDs []string
	for _, ti := range typeInstances {
		tiToUpdateIDs = append(tiToUpdateIDs, ti.ID)
	}

	// fetch details of the TypeInstances to update
	tisToUpdate, err := u.client.FindTypeInstances(ctx, tiToUpdateIDs, local.WithFields(local.TypeInstanceRootFields|local.TypeInstanceBackendFields))
	if err != nil {
		return nil, nil, errors.Wrap(err, "while fetching TypeInstance values")
	}

	// get IDs for storage backends
	tiToBackendIDMapping := map[tiToUpdateID]backendTIID{}
	var ids []string
	for _, ti := range tisToUpdate {
		if ti.Backend == nil {
			continue
		}

		tiToBackendIDMapping[tiToUpdateID(ti.ID)] = backendTIID(ti.Backend.ID)
		ids = append(ids, ti.Backend.ID)
	}

	backendValues, err := resolveBackendsValues(ctx, u.client, ids)
	if err != nil {
		return nil, nil, err
	}

	return backendValues, tiToBackendIDMapping, nil
}

func (u *Update) render(payload []graphqllocal.UpdateTypeInstancesInput, values map[string]map[string]interface{}, shouldIncludeValue func(tiToUpdate graphqllocal.UpdateTypeInstancesInput) (bool, error)) error {
	for _, typeInstance := range payload {
		value, ok := values[typeInstance.ID]
		if !ok {
			return ErrMissingTypeInstanceValue(typeInstance.ID)
		}

		if isTypeInstanceWithLegacySyntax(u.log, value) {
			typeInstance.TypeInstance.Value = value
			continue
		}

		data, err := json.Marshal(value)
		if err != nil {
			return errors.Wrap(err, "while marshaling TypeInstance")
		}

		unmarshalledTIValue := graphqllocal.UpdateTypeInstanceInput{}
		err = json.Unmarshal(data, &unmarshalledTIValue)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling TypeInstance")
		}

		if unmarshalledTIValue.Backend != nil {
			typeInstance.TypeInstance.Backend = unmarshalledTIValue.Backend
		}

		includeValue, err := shouldIncludeValue(typeInstance)
		if err != nil {
			return err
		}

		if !includeValue {
			u.log.Info("Skipping sending TypeInstance value", zap.String("ID", typeInstance.ID))
			continue
		}

		typeInstance.TypeInstance.Value = unmarshalledTIValue.Value
	}
	return nil
}

func (u *Update) shouldIncludeTIValueFn(backends map[string]storage_backend.TypeValue, mapping map[tiToUpdateID]backendTIID) func(tiToUpdate graphqllocal.UpdateTypeInstancesInput) (bool, error) {
	return func(tiToUpdate graphqllocal.UpdateTypeInstancesInput) (bool, error) {
		if tiToUpdate.TypeInstance == nil {
			return false, errors.New("typeInstance cannot be nil")
		}

		if tiToUpdate.TypeInstance.Backend == nil {
			return true, nil
		}

		backendID, exists := mapping[tiToUpdateID(tiToUpdate.ID)]
		if !exists {
			return false, fmt.Errorf("cannot retrieve backend ID for the TypeInstance to update with ID %q", tiToUpdate.ID)
		}

		backend, exists := backends[string(backendID)]
		if !exists {
			return false, fmt.Errorf("cannot retrieve value for the storage backend TypeInstance with ID %q", backendID)
		}

		return backend.AcceptValue, nil
	}
}

func (u *Update) updateTypeInstances(ctx context.Context, in []graphqllocal.UpdateTypeInstancesInput) ([]graphqllocal.TypeInstance, error) {
	return u.client.Local.UpdateTypeInstances(ctx, in)
}
