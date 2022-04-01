package testing

import (
	"context"
	"encoding/json"

	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/ptr"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

// StorageBackendRegisterConfig holds configuration for StorageBackendRegister.
type StorageBackendRegisterConfig struct {
	Logger                logger.Config
	LocalHubEndpoint      string `envconfig:"default=http://capact-hub-local.capact-system/graphql"`
	TestStorageBackendURL string `envconfig:"default=capact-test-storage-backend.capact-system:50051"`
}

const testStorageTypePath = "cap.type.capactio.capact.validation.storage"
const testStorageTypeRevision = "0.1.0"
const testStorageTypeContextSchema = `
	{
      "$schema": "http://json-schema.org/draft-07/schema",
      "type": "object",
      "required": [
        "provider"
      ],
      "properties": {
        "provider": {
          "$id": "#/properties/context/properties/provider",
          "type": "string",
          "enum": [
            "aws_secretsmanager",
            "dotenv"
          ]
        }
      },
      "additionalProperties": false
    }
`

type typeInstanceValue struct {
	URL           string      `json:"url"`
	AcceptValue   bool        `json:"acceptValue"`
	ContextSchema interface{} `json:"contextSchema"`
}

// StorageBackendRegister provides functionality to produce and upload test storage backend TypeInstance.
type StorageBackendRegister struct {
	logger      *zap.Logger
	localHubCli *local.Client
	cfg         StorageBackendRegisterConfig
}

// NewStorageBackendRegister returns a new StorageBackendRegister instance.
func NewStorageBackendRegister() (*StorageBackendRegister, error) {
	var cfg StorageBackendRegisterConfig
	err := envconfig.Init(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuration")
	}

	logger, err := logger.New(cfg.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "while creating zap logger")
	}

	client := local.NewDefaultClient(cfg.LocalHubEndpoint)

	return &StorageBackendRegister{
		logger:      logger,
		localHubCli: client,
		cfg:         cfg,
	}, nil
}

// RegisterTypeInstances produces and uploads TypeInstances which describe Test storage backend.
func (i *StorageBackendRegister) RegisterTypeInstances(ctx context.Context) error {
	var contextSchema interface{}
	err := json.Unmarshal([]byte(testStorageTypeContextSchema), &contextSchema)
	if err != nil {
		return errors.Wrap(err, "while unmarshaling contextSchema")
	}
	in := &hublocalgraphql.CreateTypeInstanceInput{
		CreatedBy: ptr.String("populator/test-storage-backend-registration"),
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     testStorageTypePath,
			Revision: testStorageTypeRevision,
		},
		Value: typeInstanceValue{
			URL:           i.cfg.TestStorageBackendURL,
			AcceptValue:   true,
			ContextSchema: contextSchema,
		},
	}

	id, err := i.localHubCli.CreateTypeInstance(ctx, in)
	if err != nil {
		return errors.Wrap(err, "while creating TypeInstance")
	}

	i.logger.Info("Successfully created Test Storage Backend TypeInstance", zap.String("id", id))

	return nil
}
