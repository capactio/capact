package helm

import (
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

type ChartRenderer interface {
	Do(chartData *chart.Chart, release *release.Release, additionalOutputTemplate []byte) ([]byte, error)
}

type Outputter struct {
	log      *zap.Logger
	renderer ChartRenderer
}

func NewOutputter(log *zap.Logger, renderer ChartRenderer) *Outputter {
	return &Outputter{log: log, renderer: renderer}
}

func (o *Outputter) ProduceHelmRelease(repository string, helmRelease *release.Release) ([]byte, error) {
	releaseData := ChartRelease{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Chart: Chart{
			Name:    helmRelease.Chart.Metadata.Name,
			Version: helmRelease.Chart.Metadata.Version,
			Repo:    repository,
		},
	}

	bytes, err := yaml.Marshal(&releaseData)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling yaml")
	}

	return bytes, nil
}

// TODO: consider to get rid of the chrt arg and use rel.Chart instead.
func (o *Outputter) ProduceAdditional(args OutputArgs, chrt *chart.Chart, rel *release.Release) ([]byte, error) {
	if strings.TrimSpace(args.GoTemplate) == "" {
		o.log.Debug("No additional output to render and save. skipping...")
		return nil, nil
	}

	bytes, err := o.renderer.Do(chrt, rel, []byte(args.GoTemplate))
	if err != nil {
		return nil, errors.Wrap(err, "while rendering additional output")
	}

	return bytes, nil
}
