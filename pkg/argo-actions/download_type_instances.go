package argoactions

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

const DownloadAction = "DownloadAction"

type DownloadConfig struct {
	ID   string
	Path string
}

type Download struct {
	cfg    []DownloadConfig
	client *local.Client
}

func NewDownloadAction(client *local.Client, cfg []DownloadConfig) Action {
	return &Download{
		client: client,
		cfg:    cfg,
	}
}

func (d *Download) Do(ctx context.Context) error {
	for _, config := range d.cfg {
		typeInstance, err := d.client.GetTypeInstance(context.TODO(), config.ID)
		if err != nil {
			return err
		}
		if typeInstance == nil {
			return fmt.Errorf("failed to find TypeInstance %s", config.ID)
		}

		data, err := yaml.Marshal(typeInstance.Spec.Value)
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
