package argoactions

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

// UploadAction represents the upload TypeInstances action.
const UploadAction = "UploadAction"

// UploadConfig stores the configuration parameters for the upload TypeInstances action.
type UploadConfig struct {
	PayloadFilepath  string
	TypeInstancesDir string
}

// Upload implements the Action interface.
// It is used to upload TypeInstances to the Local Hub.
type Upload struct {
	log    *zap.Logger
	client *hubclient.Client
	cfg    UploadConfig
}

// NewUploadAction returns a new Upload instance.
func NewUploadAction(log *zap.Logger, client *hubclient.Client, cfg UploadConfig) Action {
	return &Upload{
		log:    log,
		client: client,
		cfg:    cfg,
	}
}

// Do uploads TypeInstances to the Local Hub.
func (u *Upload) Do(ctx context.Context) error {
	payloadBytes, err := ioutil.ReadFile(u.cfg.PayloadFilepath)
	if err != nil {
		return errors.Wrap(err, "while reading payload file")
	}

	payload := &graphqllocal.CreateTypeInstancesInput{}
	if err := yaml.Unmarshal(payloadBytes, payload); err != nil {
		return errors.Wrap(err, "while unmarshaling payload bytes")
	}

	if len(payload.TypeInstances) == 0 {
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
		return errors.Wrap(err, "while rendering CreateTypeInstancesInput")
	}

	u.log.Info("Validating TypeInstances")

	r := validation.ResultAggregator{}
	err = r.Report(validation.ValidateTypeInstancesToCreate(ctx, u.client, payload))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}
	if r.ErrorOrNil() != nil {
		return r.ErrorOrNil()
	}

	u.log.Info("Uploading TypeInstances to Hub...", zap.Int("TypeInstance count", len(payload.TypeInstances)))

	uploadOutput, err := u.uploadTypeInstances(ctx, payload)
	if err != nil {
		return errors.Wrap(err, "while uploading TypeInstances")
	}

	for _, ti := range uploadOutput {
		u.log.Info("TypeInstance uploaded", zap.String("alias", ti.Alias), zap.String("ID", ti.ID))
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

		// render value
		if _, ok = value["value"]; ok {
			typeInstanceValue, ok := value["value"].(map[string]interface{})
			if !ok {
				return ErrTypeConversion("value", "map[string]interface{}", *typeInstance.Alias)
			}
			typeInstance.Value = typeInstanceValue
		} else {
			// for backward compatibility, if there is an artifact without value/backend syntax,
			// treat it as a value for TypeInstance
			typeInstance.Value = value
			continue
		}

		// render backend
		if _, ok := value["backend"]; ok {
			typeInstanceBackend, ok := value["backend"].(map[string]interface{})
			if !ok {
				return ErrTypeConversion("backend", "map[string]interface{}", *typeInstance.Alias)
			}
			backendContext, ok := typeInstanceBackend["context"].(map[string]interface{})
			if !ok {
				return ErrTypeConversion("backend.context", "map[string]interface{}", *typeInstance.Alias)
			}
			typeInstance.Backend.Context = backendContext
		}
	}
	return nil
}

func (u *Upload) uploadTypeInstances(ctx context.Context, in *graphqllocal.CreateTypeInstancesInput) ([]graphqllocal.CreateTypeInstanceOutput, error) {
	return u.client.Local.CreateTypeInstances(ctx, in)
}
