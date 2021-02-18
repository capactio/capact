package argoactions

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	graphqllocal "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"sigs.k8s.io/yaml"
)

const UploadAction = "UploadAction"

type UploadConfig struct {
	PayloadFilepath  string
	TypeInstancesDir string
}

type Upload struct {
	client *local.Client
	cfg    UploadConfig
}

func ErrMissingTypeInstanceValue(typeInstanceName string) error {
	return errors.Errorf("missing value for TypeInstance %s", typeInstanceName)
}

func NewUploadAction(client *local.Client, cfg UploadConfig) Action {
	return &Upload{
		client: client,
		cfg:    cfg,
	}
}

func (u *Upload) Do(ctx context.Context) error {
	payloadBytes, err := ioutil.ReadFile(u.cfg.PayloadFilepath)
	if err != nil {
		return errors.Wrap(err, "while reading payload file")
	}

	payload := &graphqllocal.CreateTypeInstancesInput{}
	if err := yaml.Unmarshal(payloadBytes, payload); err != nil {
		return errors.Wrap(err, "while unmarshaling payload bytes")
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

	if err := u.render(payload, typeInstanceValues); err != nil {
		return errors.Wrap(err, "while rendering CreateTypeInstancesInput")
	}

	if err := u.uploadTypeInstances(ctx, payload); err != nil {
		return errors.Wrap(err, "while uploading TypeInstances")
	}

	return nil
}

func (u *Upload) render(payload *graphqllocal.CreateTypeInstancesInput, values map[string]map[string]interface{}) error {
	for i := range payload.TypeInstances {
		typeInstance := payload.TypeInstances[i]

		value, ok := values[*typeInstance.Alias]
		if !ok {
			return ErrMissingTypeInstanceValue(*typeInstance.Alias)
		}

		typeInstance.Value = value
	}
	return nil
}

func (u *Upload) uploadTypeInstances(ctx context.Context, in *graphqllocal.CreateTypeInstancesInput) error {
	_, err := u.client.CreateTypeInstances(ctx, in)
	return err
}
