package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type installer struct {
	actionCfgProducer   actionConfigProducer
	log                 *zap.Logger
	out                 outputter
	repositoryCachePath string
}

func newInstaller(log *zap.Logger, repositoryCachePath string, actionCfgProducer actionConfigProducer, outputter outputter) *installer {
	return &installer{
		log:                 log,
		actionCfgProducer:   actionCfgProducer,
		repositoryCachePath: repositoryCachePath,
		out:                 outputter,
	}
}

func (i *installer) Do(_ context.Context, in Input) (Output, Status, error) {
	actCfg, err := i.actionCfgProducer(in.Ctx.Platform.Namespace)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while creating Helm action config")
	}

	installCli := i.initActionInstallFromInput(actCfg, in)

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

	values, err := readValueOverrides(in.Args.Values, in.Args.ValuesFromFile)
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

	releaseOut, err := i.out.ProduceHelmRelease(in.Args.Chart.Repo, helmRelease)
	if err != nil {
		return Output{}, Status{}, errors.Wrap(err, "while saving default output")
	}

	additionalOut, err := i.out.ProduceAdditional(in.Args.Output, chartData, helmRelease)
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

func (i *installer) initActionInstallFromInput(cfg *action.Configuration, in Input) *action.Install {
	installCli := action.NewInstall(cfg)
	installCli.Wait = true

	// context
	installCli.DryRun = in.Ctx.DryRun
	installCli.Namespace = in.Ctx.Platform.Namespace
	installCli.Timeout = in.Ctx.Timeout.Duration()

	// common args
	installCli.DisableHooks = in.Args.CommonArgs.NoHooks
	installCli.ChartPathOptions.Version = in.Args.CommonArgs.Chart.Version
	installCli.ChartPathOptions.RepoURL = in.Args.CommonArgs.Chart.Repo

	// install args
	installCli.GenerateName = in.Args.InstallArgs.GenerateName
	installCli.Replace = in.Args.InstallArgs.Replace

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
