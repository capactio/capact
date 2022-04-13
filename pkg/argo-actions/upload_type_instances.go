package argoactions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	storage_backend "capact.io/capact/pkg/hub/storage-backend"

	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

const (
	// UploadAction represents the upload TypeInstances action.
	UploadAction = "UploadAction"
	valueKey     = "value"
	backendKey   = "backend"
)

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

//UploadTypeInstanceData represents the TypeInstance data to upload.
type UploadTypeInstanceData struct {
	Value   interface{} `json:"value"`
	Backend *Backend    `json:"backend,omitempty"`
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

	backends, err := u.resolveBackendsValues(ctx, payload.TypeInstances)
	if err != nil {
		return errors.Wrap(err, "while resolving storage backends values")
	}

	if err := u.render(payload, typeInstanceValues, u.shouldIncludeTIValueFn(backends)); err != nil {
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

func (u *Upload) resolveBackendsValues(ctx context.Context, typeInstances []*graphqllocal.CreateTypeInstanceInput) (map[string]storage_backend.TypeInstanceValue, error) {
	// get IDs
	var ids []string
	for _, ti := range typeInstances {
		if ti.Backend == nil {
			continue
		}

		ids = append(ids, ti.Backend.ID)
	}

	return resolveBackendsValues(ctx, u.client, ids)
}

func (u *Upload) render(payload *graphqllocal.CreateTypeInstancesInput, values map[string]map[string]interface{}, shouldIncludeValue func(tiToUpload graphqllocal.CreateTypeInstanceInput) (bool, error)) error {
	for i := range payload.TypeInstances {
		typeInstance := payload.TypeInstances[i]

		value, ok := values[*typeInstance.Alias]
		if !ok {
			return ErrMissingTypeInstanceValue(*typeInstance.Alias)
		}

		if isTypeInstanceWithLegacySyntax(u.log, value) {
			typeInstance.Value = value
			continue
		}

		data, err := json.Marshal(value)
		if err != nil {
			return errors.Wrap(err, "while marshaling TypeInstance")
		}

		unmarshalledTIValue := UploadTypeInstanceData{}
		err = json.Unmarshal(data, &unmarshalledTIValue)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling TypeInstance")
		}

		// check
		if unmarshalledTIValue.Backend != nil {
			if typeInstance.Backend == nil {
				typeInstance.Backend = &graphqllocal.TypeInstanceBackendInput{}
			}

			typeInstance.Backend.Context = unmarshalledTIValue.Backend.Context
		}

		includeValue, err := shouldIncludeValue(*typeInstance)
		if err != nil {
			return err
		}

		if !includeValue {
			u.log.Info("Skipping sending TypeInstance value", zap.Stringp("Alias", typeInstance.Alias), zap.String("Backend ID", typeInstance.Backend.ID))
			continue
		}

		typeInstance.Value = unmarshalledTIValue.Value
	}
	return nil
}

func (u *Upload) uploadTypeInstances(ctx context.Context, in *graphqllocal.CreateTypeInstancesInput) ([]graphqllocal.CreateTypeInstanceOutput, error) {
	return u.client.Local.CreateTypeInstances(ctx, in)
}

func (u *Upload) shouldIncludeTIValueFn(backends map[string]storage_backend.TypeInstanceValue) func(tiToCreate graphqllocal.CreateTypeInstanceInput) (bool, error) {
	return func(tiToCreate graphqllocal.CreateTypeInstanceInput) (bool, error) {
		if tiToCreate.Backend == nil || tiToCreate.Backend.ID == "" {
			return true, nil
		}

		backend, exists := backends[tiToCreate.Backend.ID]
		if !exists {
			return false, fmt.Errorf("cannot retrieve value for the storage backend TypeInstance with ID %q", tiToCreate.Backend.ID)
		}

		return backend.AcceptValue, nil
	}
}

func isTypeInstanceWithLegacySyntax(logger *zap.Logger, value map[string]interface{}) bool {
	_, hasValue := value[valueKey]
	_, hasBackend := value[backendKey]
	if !hasValue && !hasBackend {
		// for backward compatibility, if there is an artifact without value/backend syntax,
		// treat it as a value for TypeInstance
		logger.Info(fmt.Sprintf("Found legacy TypeInstance syntax without '%s' and '%s'", valueKey, backendKey))
		return true
	}
	logger.Info(fmt.Sprintf("Processing TypeInstance '%s' and '%s'", valueKey, backendKey), zap.Bool("hasValue", hasValue), zap.Bool("hasBackend", hasBackend))
	return false
}
