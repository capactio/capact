package argoactions

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	graphqllocal "projectvoltron.dev/voltron/pkg/och/api/graphql/local-v2"
	local "projectvoltron.dev/voltron/pkg/och/client/local/v2"
	"sigs.k8s.io/yaml"
)

const UpdateAction = "UpdateAction"

type UpdateConfig struct {
	PayloadFilepath  string
	TypeInstancesDir string
}

type Update struct {
	log    *zap.Logger
	client *local.Client
	cfg    UpdateConfig
}

func NewUpdateAction(log *zap.Logger, client *local.Client, cfg UpdateConfig) Action {
	return &Update{
		log:    log,
		client: client,
		cfg:    cfg,
	}
}

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
		u.log.Info("No TypeInstances to upload")
		return nil
	}

	files, err := ioutil.ReadDir(u.cfg.TypeInstancesDir)
	if err != nil {
		return errors.Wrap(err, "while listing Type Instances directory")
	}

	typeInstanceValues := map[string]map[string]interface{}{}

	for _, f := range files {
		path := fmt.Sprintf("%s/%s", u.cfg.TypeInstancesDir, f.Name())

		typeInstanceValueBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "while reading TypeInstance value file %s", path)
		}

		values := map[string]interface{}{}
		if err := yaml.Unmarshal(typeInstanceValueBytes, &values); err != nil {
			return errors.Wrapf(err, "while unmarshaling bytes from %s file", path)
		}

		typeInstanceValues[f.Name()] = values
	}

	if err := u.render(ctx, payload, typeInstanceValues); err != nil {
		return errors.Wrap(err, "while rendering UpdateTypeInstancesInput")
	}

	u.log.Info("Updating TypeInstances in OCH...", zap.Int("TypeInstance count", len(payload)))

	uploadOutput, err := u.updateTypeInstances(ctx, payload)
	if err != nil {
		return errors.Wrap(err, "while updating TypeInstances")
	}

	for _, ti := range uploadOutput {
		u.log.Info("TypeInstance updated", zap.String("ID", ti.ID))
	}

	return nil
}

func (u *Update) render(ctx context.Context, payload []graphqllocal.UpdateTypeInstancesInput, values map[string]map[string]interface{}) error {
	for i := range payload {
		typeInstance := payload[i]

		value, ok := values[typeInstance.ID]
		if !ok {
			return ErrMissingTypeInstanceValue(typeInstance.ID)
		}

		typeInstance.TypeInstance.Value = value

		resourceVersion, err := u.fetchTypeInstanceResourceVersion(ctx, typeInstance.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting resourceVersion for TypeInstance %s", typeInstance.ID)
		}
		typeInstance.TypeInstance.ResourceVersion = resourceVersion
	}
	return nil
}

func (u *Update) fetchTypeInstanceResourceVersion(ctx context.Context, id string) (int, error) {
	ti, err := u.client.FindTypeInstance(ctx, id)
	if err != nil {
		return 0, errors.Wrap(err, "while finding TypeInstance")
	}

	if ti == nil || ti.LatestResourceVersion == nil {
		return 0, ErrMissingResourceVersion()
	}

	return ti.LatestResourceVersion.ResourceVersion, nil
}

func (u *Update) updateTypeInstances(ctx context.Context, in []graphqllocal.UpdateTypeInstancesInput) ([]graphqllocal.TypeInstance, error) {
	return u.client.UpdateTypeInstances(ctx, in)
}
