package argoactions

import (
	"context"
	"fmt"

	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/runner"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
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
		typeInstance, err := d.client.FindTypeInstance(ctx, config.ID)
		if err != nil {
			return err
		}
		if typeInstance == nil {
			return fmt.Errorf("failed to find TypeInstance with ID %q", config.ID)
		}

		data, err := yaml.Marshal(typeInstance.LatestResourceVersion.Spec.Value)
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
