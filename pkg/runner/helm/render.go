package helm

import (
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
)

const additionalOutputTemplateName = "additionalOutputTemplate"

type renderEngine interface {
	Render(*chart.Chart, chartutil.Values) (map[string]string, error)
}

type helmRenderer struct {
	renderEngine renderEngine
}

func newHelmRenderer(renderEngine renderEngine) *helmRenderer {
	return &helmRenderer{renderEngine: renderEngine}
}

func (r *helmRenderer) Do(chartData *chart.Chart, release *release.Release, additionalOutputTemplate []byte) ([]byte, error) {
	chartData.Templates = append(chartData.Templates, &chart.File{
		Name: additionalOutputTemplateName,
		Data: additionalOutputTemplate,
	})

	caps := chartutil.DefaultCapabilities
	releaseOptions := chartutil.ReleaseOptions{
		Name:      release.Name,
		Namespace: release.Namespace,
		Revision:  release.Version,
		IsInstall: true,
	}
	values, err := chartutil.ToRenderValues(chartData, release.Config, releaseOptions, caps)
	if err != nil {
		return nil, errors.Wrap(err, "while merging values")
	}

	files, err := r.renderEngine.Render(chartData, values)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering chart")
	}

	path := fmt.Sprintf("%s/%s", chartData.Metadata.Name, additionalOutputTemplateName)
	rendered, exists := files[path]
	if !exists {
		return nil, fmt.Errorf("rendered file '%v' doesnt exist", additionalOutputTemplateName)
	}

	return []byte(rendered), nil
}
