package argoactions

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/local"
	"capact.io/capact/pkg/runner"
)

// DownloadAction represents the download TypeInstance action.
const DownloadAction = "DownloadAction"

// DownloadConfig stores the configuration parameters for the download TypeInstance action.
type DownloadConfig struct {
	ID   string
	Path string
}

// Download implements the Action interface.
// It is used to download a TypeInstance from the Local Hub and save on local filesystem.
type Download struct {
	log    *zap.Logger
	cfg    []DownloadConfig
	client *hubclient.Client
}

// DownloadTypeInstanceData represents the TypeInstance data to download.
type DownloadTypeInstanceData struct {
	Value   interface{} `json:"value"`
	Backend *Backend    `json:"backend,omitempty"`
}

// Backend represents the TypeInstance Backend.
type Backend struct {
	Context interface{} `json:"context,omitempty"`
}

// NewDownloadAction returns a new Download instance.
func NewDownloadAction(log *zap.Logger, client *hubclient.Client, cfg []DownloadConfig) Action {
	return &Download{
		log:    log,
		cfg:    cfg,
		client: client,
	}
}

// Do downloads a TypeInstance from the Local Hub.
func (d *Download) Do(ctx context.Context) error {
	for _, config := range d.cfg {
		d.log.Info("Downloading TypeInstance", zap.String("ID", config.ID), zap.String("Path", config.Path))
		typeInstance, err := d.client.FindTypeInstance(ctx, config.ID, local.WithFields(local.TypeInstanceLatestResourceVersionFields))
		if err != nil {
			return err
		}
		if typeInstance == nil {
			return fmt.Errorf("failed to find TypeInstance with ID %q", config.ID)
		}

		typeInstanceData := DownloadTypeInstanceData{
			Value: typeInstance.LatestResourceVersion.Spec.Value,
		}

		if typeInstance.LatestResourceVersion.Spec.Backend != nil &&
			typeInstance.LatestResourceVersion.Spec.Backend.Context != nil {
			typeInstanceData.Backend = &Backend{
				Context: typeInstance.LatestResourceVersion.Spec.Backend.Context,
			}
		}

		data, err := yaml.Marshal(typeInstanceData)
		if err != nil {
			return errors.Wrap(err, "while marshaling TypeInstance to YAML")
		}
		err = runner.SaveToFile(config.Path, data)
		if err != nil {
			return errors.Wrapf(err, "while saving TypeInstance(%s) to file %s", config.ID, config.Path)
		}
	}
	return nil
}
