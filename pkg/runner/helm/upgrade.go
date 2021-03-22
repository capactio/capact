package helm

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/action"
)

type upgrader struct {
	actionCfg           *action.Configuration
	log                 *zap.Logger
	out                 outputter
	repositoryCachePath string
	helmReleasePath     string
}

func newUpgrader(log *zap.Logger, repositoryCachePath string, helmReleasePath string, actionCfg *action.Configuration, outputter outputter) *upgrader {
	return &upgrader{
		log:                 log,
		actionCfg:           actionCfg,
		out:                 outputter,
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
		return Output{}, Status{}, fmt.Errorf(
			"namespace from Runner Context (%q) differs with the Helm Release namespace (%q)",
			in.Ctx.Platform.Namespace,
			helmReleaseData.Namespace,
		)
	}

	helmChart := i.mergeHelmChartData(helmReleaseData, in)

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

	values, err := readValueOverrides(in.Args.Values, in.Args.ValuesFromFile)
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

	releaseOut, err := i.out.ProduceHelmRelease(helmChart.Repo, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while saving default output")
	}

	additionalOut, err := i.out.ProduceAdditional(in.Args, chartData, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while rendering and saving additional output")
	}

	status := Status{
		Succeeded: true,
		Message:   fmt.Sprintf("release %q upgraded successfully in namespace %q", helmRelease.Name, helmRelease.Namespace),
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

func (i *upgrader) mergeHelmChartData(helmRelease ChartRelease, in Input) Chart {
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
	upgradeCli.Wait = true

	// context
	upgradeCli.DryRun = in.Ctx.DryRun
	upgradeCli.Namespace = in.Ctx.Platform.Namespace
	upgradeCli.Timeout = in.Ctx.Timeout.Duration()

	// common args
	upgradeCli.DisableHooks = in.Args.CommonArgs.NoHooks
	upgradeCli.ChartPathOptions.Version = in.Args.CommonArgs.Chart.Version
	upgradeCli.ChartPathOptions.RepoURL = in.Args.CommonArgs.Chart.Repo

	// upgrade args
	upgradeCli.ReuseValues = in.Args.UpgradeArgs.ReuseValues
	upgradeCli.ResetValues = in.Args.UpgradeArgs.ResetValues

	// helm chart args
	upgradeCli.ChartPathOptions.Version = helmChart.Version
	upgradeCli.ChartPathOptions.RepoURL = helmChart.Repo

	return upgradeCli
}
