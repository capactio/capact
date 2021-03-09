package argoactions

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"projectvoltron.dev/voltron/pkg/och/client/local/v2"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

const DownloadAction = "DownloadAction"

type DownloadConfig struct {
	ID   string
	Path string
}

type Download struct {
	log    *zap.Logger
	cfg    []DownloadConfig
	client *local.Client
}

func NewDownloadAction(log *zap.Logger, client *local.Client, cfg []DownloadConfig) Action {
	return &Download{
		log:    log,
		cfg:    cfg,
		client: client,
	}
}

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
