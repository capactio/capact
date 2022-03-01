package argoactions

import (
	"context"
	"io/ioutil"
	"path"
	"path/filepath"

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

	payload := []graphqllocal.UpdateTypeInstancesInput{}
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

	if err := u.render(payload, typeInstanceValues); err != nil {
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

func (u *Update) render(payload []graphqllocal.UpdateTypeInstancesInput, values map[string]map[string]interface{}) error {
	for _, typeInstance := range payload {
		value, ok := values[typeInstance.ID]
		if !ok {
			return ErrMissingTypeInstanceValue(typeInstance.ID)
		}

		typeInstance.TypeInstance.Value = value
	}
	return nil
}

func (u *Update) updateTypeInstances(ctx context.Context, in []graphqllocal.UpdateTypeInstancesInput) ([]graphqllocal.TypeInstance, error) {
	return u.client.Local.UpdateTypeInstances(ctx, in)
}
