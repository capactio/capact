package helm

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io/ioutil"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/action"
)

type upgrader struct {
	actionCfg           *action.Configuration
	log                 *zap.Logger
	renderer            renderer
	repositoryCachePath string
	helmReleasePath     string
}

func newUpgrader(log *zap.Logger, repositoryCachePath string, helmReleasePath string, actionCfg *action.Configuration, renderer renderer) *upgrader {
	return &upgrader{
		log:                 log,
		actionCfg:           actionCfg,
		renderer:            renderer,
		repositoryCachePath: repositoryCachePath,
		helmReleasePath:     helmReleasePath,
	}
}

func (i *upgrader) Do(_ context.Context, in Input) (Output, Status, error) {
	if i.helmReleasePath == "" {
		return Output{}, Status{}, errors.New("path to Helm Release is required for upgrade")
	}

	helmReleaseData, err := i.loadHelmReleaseData(i.helmReleasePath)
	if err != nil {
		return Output{}, Status{}, err
	}

	if in.Ctx.Platform.Namespace != helmReleaseData.Namespace {
		return Output{}, Status{}, fmt.Errorf("namespace from Runner Context (%q) differs with the Helm Release namespace (%q)", in.Ctx.Platform.Namespace, helmReleaseData.Namespace)
	}

	helmChart := i.getHelmChartData(in, helmReleaseData)

	upgradeCli := i.initActionUpgradeFromInput(in, helmChart)

	chartPath, err := upgradeCli.ChartPathOptions.LocateChart(helmChart.Name, &cli.EnvSettings{
		RepositoryCache: i.repositoryCachePath,
	})
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while locating Helm chart")
	}

	chartData, err := loader.Load(chartPath)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while loading Helm chart")
	}

	values, err := i.getValues(in.Args.Values, in.Args.ValuesFromFile)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while reading values")
	}

	helmRelease, err := upgradeCli.Run(helmReleaseData.Name, chartData, values)

	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while installing Helm chart")
	}

	if helmRelease == nil {
		return Output{}, Status{}, errors.Wrap(err, "Helm release is nil")
	}

	releaseOut, err := i.releaseOutputFrom(in.Args, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while saving default output")
	}

	additionalOut, err := i.additionalOutputFrom(in.Args, chartData, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while rendering and saving additional output")
	}

	status := Status{
		Succeeded: true,
		Message:   fmt.Sprintf("release %q installed successfully in namespace %q", helmRelease.Name, helmRelease.Namespace),
	}

	return Output{
		Release:    releaseOut,
		Additional: additionalOut,
	}, status, nil
}

func (i *upgrader) loadHelmReleaseData(path string) (ChartRelease, error) {
	i.log.Debug("Reading Helm Release data from file", zap.String("path", path))
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return ChartRelease{}, errors.Wrapf(err, "while reading values from file %q", path)
	}

	var chartRelease ChartRelease
	if err := yaml.Unmarshal(bytes, &chartRelease); err != nil {
		return ChartRelease{}, errors.Wrapf(err, "while parsing %q", path)
	}
	return chartRelease, nil
}

func (i *upgrader) getHelmChartData(in Input, helmRelease ChartRelease) Chart {
	helmChart := helmRelease.Chart

	if in.Args.Chart.Name != "" {
		helmChart.Name = in.Args.Chart.Name
	}

	if in.Args.Chart.Repo != "" {
		helmChart.Repo = in.Args.Chart.Repo
	}

	if in.Args.Chart.Version != "" {
		helmChart.Version = in.Args.Chart.Version
	}

	return helmChart
}

func (i *upgrader) initActionUpgradeFromInput(in Input, helmChart Chart) *action.Upgrade {
	upgradeCli := action.NewUpgrade(i.actionCfg)
	upgradeCli.DryRun = in.Ctx.DryRun
	upgradeCli.Namespace = in.Ctx.Platform.Namespace
	upgradeCli.Wait = true
	upgradeCli.Timeout = in.Ctx.Timeout.Duration()

	upgradeCli.DisableHooks = in.Args.CommonArgs.NoHooks
	upgradeCli.ChartPathOptions.Version = in.Args.CommonArgs.Chart.Version
	upgradeCli.ChartPathOptions.RepoURL = in.Args.CommonArgs.Chart.Repo

	upgradeCli.ReuseValues = in.Args.UpgradeArgs.ReuseValues
	upgradeCli.ResetValues = in.Args.UpgradeArgs.ResetValues

	upgradeCli.ChartPathOptions.Version = helmChart.Version
	upgradeCli.ChartPathOptions.RepoURL = helmChart.Repo

	return upgradeCli
}

func (i *upgrader) getValues(inlineValues map[string]interface{}, valuesFilePath string) (map[string]interface{}, error) {
	var values map[string]interface{}

	if valuesFilePath == "" {
		return inlineValues, nil
	}

	if len(inlineValues) > 0 && valuesFilePath != "" {
		return nil, errors.New("providing values both inline and from file is currently unsupported")
	}

	i.log.Debug("Reading values from file", zap.String("path", valuesFilePath))
	bytes, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading values from file %q", valuesFilePath)
	}
	if err := yaml.Unmarshal(bytes, &values); err != nil {
		return nil, errors.Wrapf(err, "while parsing %q", valuesFilePath)
	}
	return values, nil
}

func (i *upgrader) releaseOutputFrom(args Arguments, helmRelease *release.Release) ([]byte, error) {
	releaseData := ChartRelease{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Chart: Chart{
			Name:    helmRelease.Chart.Metadata.Name,
			Version: helmRelease.Chart.Metadata.Version,
			Repo:    args.Chart.Repo,
		},
	}

	bytes, err := yaml.Marshal(&releaseData)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling yaml")
	}

	return bytes, nil
}

func (i *upgrader) additionalOutputFrom(args Arguments, chrt *chart.Chart, rel *release.Release) ([]byte, error) {
	if args.Output.GoTemplate == nil {
		i.log.Debug("No additional output to render and save. skipping...")
		return nil, nil
	}

	// yaml.Unmarshal converts YAML to JSON then uses JSON to unmarshal into an object
	// but the GoTemplate is defined via YAML, so we need to revert that change
	artifactTemplate, err := yaml.JSONToYAML(args.Output.GoTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GoTemplate property from JSON to YAML")
	}

	bytes, err := i.renderer.Do(chrt, rel, artifactTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering additional output")
	}

	return bytes, nil
}
