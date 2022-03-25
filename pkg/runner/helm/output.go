package helm

import (
	"strings"

	"capact.io/capact/internal/ptr"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

// ChartRenderer has a Do method.
// The Do method renders the additional output from the Helm Chart release.
type ChartRenderer interface {
	Do(chartData *chart.Chart, release *release.Release, additionalOutputTemplate []byte) ([]byte, error)
}

// OutputFile defines the shape of the output file.
type OutputFile struct {
	Value   interface{}        `json:"value"`
	Backend *OutputFileBackend `json:"backend"`
}

// OutputFileHelmReleaseContext contains context data for the Helm Release storage backend.
type OutputFileHelmReleaseContext struct {
	Release

	// ChartLocation specifies Helm Chart location.
	ChartLocation string `json:"chartLocation"`
}

// OutputFileAdditionalContext contains context for the Helm template storage backend.
type OutputFileAdditionalContext struct {
	// GoTemplate specifies Go template which is used to render returned value.
	GoTemplate string `json:"goTemplate"`
	// HelmRelease specifies Helm release details against which the rendering logic should be executed.
	HelmRelease Release `json:"release"`
}

// Release holds details about Helm release.
type Release struct {
	// Name specifies Helm release name for a given request.
	Name string `json:"name"`
	// Namespace specifies in which Kubernetes Namespace Helm release is located.
	Namespace string `json:"namespace"`
	// Driver specifies drivers used for storing the Helm release.
	Driver *string `json:"driver,omitempty"`
}

// OutputFileBackend defines shape of the Backend property in output file.
type OutputFileBackend struct {
	Context interface{} `json:"context"`
}

// Outputter handles producing the runner output artifacts.
type Outputter struct {
	log      *zap.Logger
	renderer ChartRenderer
}

// NewOutputter returns a new Outputer.
func NewOutputter(log *zap.Logger, renderer ChartRenderer) *Outputter {
	return &Outputter{log: log, renderer: renderer}
}

// ProduceHelmRelease creates an output artifacts with the Helm release data.
func (o *Outputter) ProduceHelmRelease(args ReleaseOutputArgs, repository, driver string, helmRelease *release.Release) ([]byte, error) {
	outputData := OutputFile{}

	if args.UseHelmReleaseStorage {
		outputData.Backend = &OutputFileBackend{
			Context: OutputFileHelmReleaseContext{
				Release: Release{
					Name:      helmRelease.Name,
					Namespace: helmRelease.Namespace,
					Driver:    ptr.String(driver),
				},
				ChartLocation: repository,
			},
		}
	} else {
		outputData.Value = ChartRelease{
			Name:      helmRelease.Name,
			Namespace: helmRelease.Namespace,
			Chart: Chart{
				Name:    helmRelease.Chart.Metadata.Name,
				Version: helmRelease.Chart.Metadata.Version,
				Repo:    repository,
			},
		}
	}

	bytes, err := yaml.Marshal(&outputData)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling yaml")
	}

	return bytes, nil
}

// ProduceAdditional creates an output artifacts from the output template provided in the args.
// TODO: consider to get rid of the chrt arg and use rel.Chart instead.
func (o *Outputter) ProduceAdditional(args OutputArgs, chrt *chart.Chart, driver string, rel *release.Release) ([]byte, error) {
	goTemplate := args.Additional.GoTemplate
	if strings.TrimSpace(goTemplate) == "" {
		// Fallback to legacy field
		goTemplate = args.LegacyGoTemplate
	}

	if strings.TrimSpace(goTemplate) == "" {
		// still nothing - exit
		o.log.Debug("No additional output to render and save. skipping...")
		return nil, nil
	}

	outputData := OutputFile{}

	if args.Additional.UseHelmTemplateStorage {
		outputData.Backend = &OutputFileBackend{
			Context: OutputFileAdditionalContext{
				HelmRelease: Release{
					Name:      rel.Name,
					Namespace: rel.Namespace,
					Driver:    ptr.String(driver),
				},
				GoTemplate: goTemplate,
			},
		}
	} else {
		bytes, err := o.renderer.Do(chrt, rel, []byte(goTemplate))
		if err != nil {
			return nil, errors.Wrap(err, "while rendering additional output")
		}

		var unmarshalled interface{}
		err = yaml.Unmarshal(bytes, &unmarshalled)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshalling additional output bytes")
		}

		outputData.Value = unmarshalled
	}

	bytes, err := yaml.Marshal(&outputData)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling yaml")
	}

	return bytes, nil
}
