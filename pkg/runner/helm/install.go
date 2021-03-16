package helm

import (
	"context"
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

type renderer interface {
	Do(chartData *chart.Chart, release *release.Release, additionalOutputTemplate []byte) ([]byte, error)
}

type installer struct {
	actionCfg           *action.Configuration
	log                 *zap.Logger
	renderer            renderer
	repositoryCachePath string
}

func newInstaller(log *zap.Logger, repositoryCachePath string, actionCfg *action.Configuration, renderer renderer) *installer {
	return &installer{
		log:                 log,
		actionCfg:           actionCfg,
		renderer:            renderer,
		repositoryCachePath: repositoryCachePath,
	}
}

func (i *installer) Do(_ context.Context, in Input) (Output, Status, error) {
	installCli := i.initActionInstallFromInput(in)

	name, chartName, err := i.nameAndChart(installCli, in.Args.Name, in.Args.Chart.Name)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while getting release name")
	}
	installCli.ReleaseName = name

	chartPath, err := installCli.ChartPathOptions.LocateChart(chartName, &cli.EnvSettings{
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

	helmRelease, err := installCli.Run(chartData, values)
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

func (i *installer) initActionInstallFromInput(in Input) *action.Install {
	installCli := action.NewInstall(i.actionCfg)
	installCli.DryRun = in.Ctx.DryRun
	installCli.Namespace = in.Ctx.Platform.Namespace
	installCli.Wait = true
	installCli.Timeout = in.Ctx.Timeout.Duration()

	installCli.GenerateName = in.Args.InstallArgs.GenerateName
	installCli.Replace = in.Args.InstallArgs.Replace

	installCli.DisableHooks = in.Args.CommonArgs.NoHooks
	installCli.ChartPathOptions.Version = in.Args.CommonArgs.Chart.Version
	installCli.ChartPathOptions.RepoURL = in.Args.CommonArgs.Chart.Repo

	return installCli
}

func (i *installer) nameAndChart(installCli *action.Install, releaseName string, chartName string) (string, string, error) {
	var nameAndChartArgs []string
	if releaseName != "" {
		nameAndChartArgs = append(nameAndChartArgs, releaseName)
	}
	nameAndChartArgs = append(nameAndChartArgs, chartName)

	return installCli.NameAndChart(nameAndChartArgs)
}

func (i *installer) getValues(inlineValues map[string]interface{}, valuesFilePath string) (map[string]interface{}, error) {
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

func (i *installer) releaseOutputFrom(args Arguments, helmRelease *release.Release) ([]byte, error) {
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

func (i *installer) additionalOutputFrom(args Arguments, chrt *chart.Chart, rel *release.Release) ([]byte, error) {
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
