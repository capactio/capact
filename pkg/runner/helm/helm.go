package helm

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

// Config holds Runner related configuration.
type Config struct {
	HelmDriver          string `envconfig:"default=secrets"`
	RepositoryCachePath string `envconfig:"default=/tmp/helm"`
}

// Runner provides functionality to run and wait for Helm operations.
type helmRunner struct {
	cfg      Config
	helmCfg  *genericclioptions.ConfigFlags
	renderer *helmRenderer
}

func newHelmRunner(k8sCfg *rest.Config, cfg Config) *helmRunner {
	helmCfg := helmConfigFlagsFrom(k8sCfg)

	renderer := newHelmRenderer(&engine.Engine{})

	return &helmRunner{
		cfg:      cfg,
		helmCfg:  helmCfg,
		renderer: renderer,
	}
}

type Arguments struct {
	Command        string                 `json:"command"`
	Name           string                 `json:"name"`
	Chart          Chart                  `json:"chart"`
	Values         map[string]interface{} `json:"values"`
	ValuesFromFile string                 `json:"valuesFromFile"`
	NoHooks        bool                   `json:"noHooks"`
	Replace        bool                   `json:"replace"`
	GenerateName   bool                   `json:"generateName"`

	Output Output `json:"output"`
}

type Output struct {
	Directory  string           `json:"directory"`
	Default    DefaultOutput    `json:"default"`
	Additional AdditionalOutput `json:"additional"`
}

type DefaultOutput struct {
	FileName string `json:"fileName"`
}

type AdditionalOutput struct {
	FileName string `json:"fileName"`
	Value    string `json:"value"`
}

type Chart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
}

type ChartRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     Chart  `json:"chart"`
}

func (r *helmRunner) Do(ctx context.Context, in runner.StartInput) (*runner.WaitForCompletionOutput, error) {
	actionConfig := new(action.Configuration)
	ns := in.ExecCtx.Platform.Namespace
	err := actionConfig.Init(r.helmCfg, ns, r.cfg.HelmDriver, func(format string, v ...interface{}) {
		log.Printf(format, v...)
	})
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Helm configuration")
	}

	var args Arguments
	err = yaml.Unmarshal(in.Args, &args)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling runner arguments")
	}

	if args.Command != "install" {
		return nil, errors.New("Unsupported command")
	}

	installCli := action.NewInstall(actionConfig)
	installCli.DryRun = in.ExecCtx.DryRun
	installCli.Namespace = ns
	installCli.Wait = true
	installCli.Timeout = in.ExecCtx.Timeout.Duration()

	installCli.GenerateName = args.GenerateName
	installCli.Replace = args.Replace
	installCli.DisableHooks = args.NoHooks
	installCli.ChartPathOptions.Version = args.Chart.Version
	installCli.ChartPathOptions.RepoURL = args.Chart.Repo

	nameAndChartArgs := r.nameAndChartArgs(args.Name, args.Chart.Name)
	name, chrtName, err := installCli.NameAndChart(nameAndChartArgs)
	if err != nil {
		return nil, errors.Wrap(err, "while getting release name")
	}
	installCli.ReleaseName = name

	chartPath, err := installCli.ChartPathOptions.LocateChart(chrtName, &cli.EnvSettings{
		RepositoryCache: r.cfg.RepositoryCachePath,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while locating Helm chart")
	}

	chrt, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "while loading Helm chart")
	}

	values, err := r.getValues(args.Values, args.ValuesFromFile)
	if err != nil {
		return nil, errors.Wrap(err, "while reading values")
	}

	helmRelease, err := installCli.Run(chrt, values)
	if err != nil {
		return nil, errors.Wrap(err, "while installing Helm chart")
	}

	if helmRelease == nil {
		return nil, errors.Wrap(err, "Helm release is nil")
	}

	releaseData := ChartRelease{
		Name:      helmRelease.Name,
		Namespace: helmRelease.Namespace,
		Chart: Chart{
			Name:    helmRelease.Chart.Metadata.Name,
			Version: helmRelease.Chart.Metadata.Version,
			Repo:    args.Chart.Repo,
		},
	}

	err = r.saveDefaultOutput(args.Output, releaseData)
	if err != nil {
		return nil, errors.Wrap(err, "while saving default output")
	}

	err = r.renderAndSaveAdditionalOutputIfShould(args.Output, chrt, helmRelease)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering and saving output")
	}

	return &runner.WaitForCompletionOutput{
		Succeeded: true,
		Message:   fmt.Sprintf("release '%s' installed successfully in namespace '%s'", helmRelease.Name, helmRelease.Namespace),
	}, nil
}

func (r *helmRunner) Name() string {
	return "helm"
}

func (r *helmRunner) nameAndChartArgs(releaseName string, chartName string) []string {
	var nameAndChartArgs []string
	if releaseName != "" {
		nameAndChartArgs = append(nameAndChartArgs, releaseName)
	}
	nameAndChartArgs = append(nameAndChartArgs, chartName)

	return nameAndChartArgs
}

func (r *helmRunner) getValues(inlineValues map[string]interface{}, valuesFilePath string) (map[string]interface{}, error) {
	var values map[string]interface{}

	if valuesFilePath == "" {
		return inlineValues, nil
	}

	if len(inlineValues) > 0 && valuesFilePath != "" {
		return nil, errors.New("providing values both inline and from file is currently unsupported")
	}

	log.Printf("Reading values from file '%s'", valuesFilePath)
	bytes, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading values from file '%s'", valuesFilePath)
	}
	if err := yaml.Unmarshal(bytes, &values); err != nil {
		return nil, errors.Wrapf(err, "while parsing '%s'", valuesFilePath)
	}
	return values, nil
}

func (r *helmRunner) saveDefaultOutput(out Output, rel ChartRelease) error {
	filePath := fmt.Sprintf("%s/%s", out.Directory, out.Default.FileName)
	log.Printf("Saving default output to %s\n...", filePath)

	bytes, err := yaml.Marshal(&rel)
	if err != nil {
		return errors.Wrap(err, "while marshaling yaml")
	}

	return r.saveToFile(filePath, bytes)
}

func (r *helmRunner) renderAndSaveAdditionalOutputIfShould(out Output, chrt *chart.Chart, rel *release.Release) error {
	if out.Additional.FileName == "" {
		log.Println("no additional output to render and save. skipping...")
		return nil
	}

	bytes, err := r.renderer.Do(chrt, rel, []byte(out.Additional.Value))
	if err != nil {
		return errors.Wrap(err, "while rendering additional output")
	}

	filePath := fmt.Sprintf("%s/%s", out.Directory, out.Additional.FileName)
	log.Printf("Saving additional output to %s\n...", filePath)

	return r.saveToFile(filePath, bytes)
}

const defaultFilePermissions = 0644

func (r *helmRunner) saveToFile(path string, bytes []byte) error {
	err := ioutil.WriteFile(path, bytes, defaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to '%s'", path)
	}

	return nil
}

func helmConfigFlagsFrom(k8sCfg *rest.Config) *genericclioptions.ConfigFlags {
	if k8sCfg == nil {
		return nil
	}

	return &genericclioptions.ConfigFlags{
		APIServer:   &k8sCfg.Host,
		Insecure:    &k8sCfg.Insecure,
		CAFile:      &k8sCfg.CAFile,
		BearerToken: &k8sCfg.BearerToken,
	}
}
