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
	actionCfgProducer   actionConfigProducer
	log                 *zap.Logger
	out                 outputter
	repositoryCachePath string
	helmReleasePath     string
}

func newUpgrader(log *zap.Logger, repositoryCachePath string, helmReleasePath string, actionCfgProducer actionConfigProducer, outputter outputter) helmCommand {
	return &upgrader{
		log:                 log,
		actionCfgProducer:   actionCfgProducer,
		out:                 outputter,
		repositoryCachePath: repositoryCachePath,
		helmReleasePath:     helmReleasePath,
	}
}

// TODO: describe that we use the Helm Chart namespace.
// "namespace from Runner Context (%q) differs with the Helm Release namespace (%q)"
func (i *upgrader) Do(_ context.Context, in Input) (Output, Status, error) {
	if i.helmReleasePath == "" {
		return Output{}, Status{}, errors.New("path to Helm Release is required for upgrade")
	}

	helmReleaseData, err := i.loadHelmReleaseData(i.helmReleasePath)
	if err != nil {
		return Output{}, Status{}, err
	}

	actCfg, err := i.actionCfgProducer(helmReleaseData.Namespace)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while creating Helm action config")
	}

	helmChartRel := i.mergeHelmChartData(helmReleaseData, in)

	upgradeCli := i.initActionUpgradeFromInput(actCfg, in, helmChartRel)

	chartPath, err := upgradeCli.ChartPathOptions.LocateChart(helmChartRel.Chart.Name, &cli.EnvSettings{
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

	helmRelease, err := upgradeCli.Run(helmChartRel.Name, chartData, values)

	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while installing Helm chart")
	}

	if helmRelease == nil {
		return Output{}, Status{}, errors.Wrap(err, "Helm release is nil")
	}

	releaseOut, err := i.out.ProduceHelmRelease(helmChartRel.Chart.Repo, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while saving default output")
	}

	additionalOut, err := i.out.ProduceAdditional(in.Args.Output, chartData, helmRelease)
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

func (i *upgrader) mergeHelmChartData(helmRelease ChartRelease, in Input) ChartRelease {
	if in.Args.Chart.Name != "" {
		helmRelease.Chart.Name = in.Args.Chart.Name
	}

	if in.Args.Chart.Repo != "" {
		helmRelease.Chart.Repo = in.Args.Chart.Repo
	}

	if in.Args.Chart.Version != "" {
		helmRelease.Chart.Version = in.Args.Chart.Version
	}

	return helmRelease
}

func (i *upgrader) initActionUpgradeFromInput(cfg *action.Configuration, in Input, helmChartRel ChartRelease) *action.Upgrade {
	upgradeCli := action.NewUpgrade(cfg)
	upgradeCli.Wait = true

	// context
	upgradeCli.DryRun = in.Ctx.DryRun
	upgradeCli.Timeout = in.Ctx.Timeout.Duration()

	// common args
	upgradeCli.DisableHooks = in.Args.CommonArgs.NoHooks
	upgradeCli.ChartPathOptions.Version = in.Args.CommonArgs.Chart.Version
	upgradeCli.ChartPathOptions.RepoURL = in.Args.CommonArgs.Chart.Repo

	// upgrade args
	upgradeCli.ReuseValues = in.Args.UpgradeArgs.ReuseValues
	upgradeCli.ResetValues = in.Args.UpgradeArgs.ResetValues

	// helm chart args
	upgradeCli.ChartPathOptions.Version = helmChartRel.Chart.Version
	upgradeCli.ChartPathOptions.RepoURL = helmChartRel.Chart.Repo
	upgradeCli.Namespace = helmChartRel.Namespace

	return upgradeCli
}
