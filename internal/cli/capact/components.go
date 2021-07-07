package capact

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"capact.io/capact/internal/cli/printer"

	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tools "capact.io/capact/internal"
	"capact.io/capact/internal/ptr"
)

type Component interface {
	InstallUpgrade(version string) error
	Name() string
	withOptions(*Options)
	withConfiguration(*action.Configuration)
	withWriter(io.Writer)
}

type ComponentData struct {
	ReleaseName string
	LocalPath   string
	Wait        bool

	Resources *Resources
	Overrides map[string]interface{}

	configuration *action.Configuration
	opts          *Options

	writer io.Writer
}

func (c *ComponentData) Name() string {
	return c.ReleaseName
}

func (c *ComponentData) installAction(version string) *action.Install {
	installCli := action.NewInstall(c.configuration)

	installCli.DryRun = c.opts.DryRun
	installCli.Namespace = c.opts.Namespace
	installCli.Timeout = c.opts.Timeout

	installCli.ChartPathOptions.Version = version
	installCli.ChartPathOptions.RepoURL = c.opts.Parameters.Override.HelmRepoURL

	installCli.NameTemplate = c.ReleaseName
	installCli.ReleaseName = c.ReleaseName

	installCli.Wait = c.Wait
	installCli.Replace = true

	return installCli
}

func (c *ComponentData) upgradeAction(version string) *action.Upgrade {
	upgradeAction := action.NewUpgrade(c.configuration)

	upgradeAction.DryRun = c.opts.DryRun
	upgradeAction.Namespace = c.opts.Namespace
	upgradeAction.Timeout = c.opts.Timeout

	upgradeAction.ChartPathOptions.Version = version
	upgradeAction.ChartPathOptions.RepoURL = c.opts.Parameters.Override.HelmRepoURL

	upgradeAction.Wait = c.Wait

	return upgradeAction
}

func (c *ComponentData) withConfiguration(configuration *action.Configuration) {
	c.configuration = configuration
}

func (c *ComponentData) withOptions(options *Options) {
	c.opts = options
}

func (c *ComponentData) withWriter(w io.Writer) {
	c.writer = w
}

func (c *ComponentData) runUpgrade(upgradeCli *action.Upgrade, values map[string]interface{}) error {
	histClient := action.NewHistory(c.configuration)
	histClient.Max = 1
	if _, err := histClient.Run(c.Name()); err == driver.ErrReleaseNotFound {
		installAction := c.installAction(upgradeCli.Version)
		return c.runInstall(installAction, values)
	}
	var chartPath string
	var err error
	var location string

	if upgradeCli.Version == LocalVersionTag {
		location = c.LocalPath
	} else {
		location = c.ReleaseName
	}

	chartPath, err = upgradeCli.ChartPathOptions.LocateChart(location, &cli.EnvSettings{
		RepositoryCache: RepositoryCache,
	})
	if err != nil {
		return errors.Wrap(err, "while locating Helm chart")
	}

	chartData, err := loader.Load(chartPath)
	if err != nil {
		return errors.Wrap(err, "while loading Helm chart")
	}

	r, err := upgradeCli.Run(c.Name(), chartData, values)
	if err != nil {
		return errors.Wrapf(err, "while upgrading Helm chart [%s]", c.Name())
	}
	c.WriteStatus(c.writer, r)
	return nil
}

func (c *ComponentData) runInstall(installCli *action.Install, values map[string]interface{}) error {
	var chartPath string
	var err error
	var location string

	if installCli.Version == LocalVersionTag {
		location = c.LocalPath
	} else {
		location = c.ReleaseName
	}

	chartPath, err = installCli.ChartPathOptions.LocateChart(location, &cli.EnvSettings{
		RepositoryCache: RepositoryCache,
	})
	if err != nil {
		return errors.Wrap(err, "while locating Helm chart")
	}

	chartData, err := loader.Load(chartPath)
	if err != nil {
		return errors.Wrap(err, "while loading Helm chart")
	}

	r, err := installCli.Run(chartData, values)
	if err != nil {
		return errors.Wrapf(err, "while installing Helm chart [%s]", installCli.ReleaseName)
	}
	c.WriteStatus(c.writer, r)

	return nil
}

// based on https://github.com/helm/helm/blob/main/cmd/helm/status.go#L112
func (c ComponentData) WriteStatus(out io.Writer, r *release.Release) {
	if r == nil {
		return
	}
	fmt.Fprint(out, "\n\n")
	fmt.Fprintf(out, "NAME: %s\n", r.Name)
	if !r.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", r.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", r.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", r.Info.Status.String())
	fmt.Fprintf(out, "REVISION: %d\n", r.Version)
	fmt.Fprintf(out, "DESCRIPTION: %s\n", r.Info.Description)
	fmt.Fprint(out, "\n")
}

var components = []Component{
	&Neo4j{
		ComponentData{
			ReleaseName: "neo4j",
			LocalPath:   path.Join(LocalChartsPath, "neo4j"),
			Wait:        true,
		},
	},
	&IngressController{
		ComponentData{
			ReleaseName: "ingress-controller",
			LocalPath:   path.Join(LocalChartsPath, "ingress-nginx"),
			Wait:        true,
		},
	},
	&Argo{
		ComponentData{
			ReleaseName: "argo",
			LocalPath:   path.Join(LocalChartsPath, "argo"),
		},
	},
	&CertManager{
		ComponentData{
			ReleaseName: "cert-manager",
			LocalPath:   path.Join(LocalChartsPath, "cert-manager"),
			Wait:        true,
		},
	},
	&Kubed{
		ComponentData{
			ReleaseName: "kubed",
			LocalPath:   path.Join(LocalChartsPath, "kubed"),
		},
	},
	&Monitoring{
		ComponentData{
			ReleaseName: "monitoring",
			LocalPath:   path.Join(LocalChartsPath, "monitoring"),
		},
	},
	&Capact{
		ComponentData{
			ReleaseName: "capact",
			LocalPath:   path.Join(LocalChartsPath, "capact"),
			Wait:        true,
		},
	},
}

type Neo4j struct {
	ComponentData
}

type IngressController struct {
	ComponentData
}

type Argo struct {
	ComponentData
}

type CertManager struct {
	ComponentData
}

type Kubed struct {
	ComponentData
}

type Monitoring struct {
	ComponentData
}

type Capact struct {
	ComponentData
}

func (n *Neo4j) InstallUpgrade(version string) error {
	upgradeCli := n.upgradeAction(version)

	values := tools.MergeMaps(map[string]interface{}{}, n.Overrides)

	return n.runUpgrade(upgradeCli, values)
}

func (a *Argo) InstallUpgrade(version string) error {
	upgradeCli := a.upgradeAction(version)

	values := tools.MergeMaps(map[string]interface{}{}, a.Overrides)

	return a.runUpgrade(upgradeCli, values)
}

func (i *IngressController) InstallUpgrade(version string) error {
	var err error
	upgradeCli := i.upgradeAction(version)

	values := map[string]interface{}{}
	if i.opts.Environment == KindEnv {
		values, err = ValuesFromString(ingressKindOverridesYaml)
		if err != nil {
			return errors.Wrap(err, "while converting override values")
		}
	} //TODO eks
	return i.runUpgrade(upgradeCli, values)
}

func (c *CertManager) InstallUpgrade(version string) error {
	upgradeCli := c.upgradeAction(version)

	values := map[string]interface{}{}

	err := c.runUpgrade(upgradeCli, values)
	if err != nil {
		return errors.Wrap(err, "while installing cert-manager")
	}

	if c.opts.Environment != KindEnv {
		return nil
	}

	// TODO if h.opts.Environment == "eks" {}

	restConfig, err := c.configuration.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "while getting k8s REST config")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: certManagerSecretName,
		},
		Data: map[string][]byte{
			"tls.crt": []byte(tlsCrt),
			"tls.key": []byte(tlsKey),
		},
	}
	err = CreateUpdateSecret(restConfig, secret, c.opts.Namespace)
	if err != nil {
		return errors.Wrapf(err, "while creating %s Secret", certManagerSecretName)
	}

	// Not using cert-manager types as it's conflicting with argo deps
	issuer := fmt.Sprintf(issuerTemplate, clusterIssuerName, certManagerSecretName)
	err = createObject(c.configuration, []byte(issuer))
	if err != nil {
		return errors.Wrapf(err, "while creating %s ClusterIssuer", clusterIssuerName)
	}
	return nil
}

func (k *Kubed) InstallUpgrade(version string) error {
	restConfig, err := k.configuration.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "while getting k8s REST config")
	}

	upgradeCli := k.upgradeAction(version)
	values := map[string]interface{}{}
	err = k.runUpgrade(upgradeCli, values)
	if err != nil {
		return errors.Wrap(err, "while running action")
	}

	err = AnnotateSecret(restConfig, "argo-minio", k.opts.Namespace, "kubed.appscode.com/sync", "")
	return errors.Wrap(err, "while annotating secret")
}

func (c *Capact) InstallUpgrade(version string) error {
	upgradeCli := c.upgradeAction(version)

	capactValues := c.opts.Parameters.Override.CapactValues
	if version == LocalVersionTag {
		capactValues.Global.ContainerRegistry.Path = LocalDockerPath
		capactValues.Global.ContainerRegistry.Tag = LocalDockerTag
	}
	s := structs.New(capactValues)
	s.TagName = "json"
	mappedValues := s.Map()

	if c.opts.Environment == KindEnv {
		values, err := ValuesFromString(capactKindOverridesYaml)
		if err != nil {
			return errors.Wrap(err, "while converting override values")
		}
		mappedValues = tools.MergeMaps(values, mappedValues)
	}
	return c.runUpgrade(upgradeCli, mappedValues)
}

func (m *Monitoring) InstallUpgrade(version string) error {
	upgradeAction := m.upgradeAction(version)

	values := map[string]interface{}{}
	return m.runUpgrade(upgradeAction, values)
}

func GetActionConfiguration(k8sCfg *rest.Config, forNamespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   &k8sCfg.Host,
		Insecure:    &k8sCfg.Insecure,
		CAFile:      &k8sCfg.CAFile,
		BearerToken: &k8sCfg.BearerToken,
		Namespace:   ptr.String(forNamespace),
	}

	debugLog := func(format string, v ...interface{}) {
		// noop
	}

	err := actionConfig.Init(helmCfg, forNamespace, "secrets", debugLog)

	if err != nil {
		return nil, errors.Wrap(err, "while initializing Helm configuration")
	}

	return actionConfig, nil
}

type Helm struct {
	configuration *action.Configuration
	opts          Options
}

func NewHelm(configuration *action.Configuration, opts Options) *Helm {
	if opts.Parameters.IncreaseResourceLimits {
		opts.Parameters.Override.CapactValues.Gateway.Resources = IncreasedGatewayResources()
		opts.Parameters.Override.CapactValues.HubPublic.Resources = IncreasedHubPublicResources()
		opts.Parameters.Override.Neo4jValues.Neo4j.Core.Resources = IncreasedNeo4jResources()
	}
	if opts.Parameters.Override.HelmRepoURL == LatestVersionTag {
		opts.Parameters.Override.HelmRepoURL = HelmRepoLatest
	}

	return &Helm{configuration: configuration, opts: opts}
}

func (h *Helm) InstallComponnents(w io.Writer, status *printer.Status) error {
	var err error
	helper := NewHelmHelper()

	for _, component := range components {
		if shouldSkipTheComponent(component.Name(), h.opts.SkipComponents) {
			continue
		}

		component.withOptions(&h.opts)
		component.withConfiguration(h.configuration)
		component.withWriter(w)

		version := h.opts.Parameters.Version
		if version == LatestVersionTag {
			version, err = helper.GetLatestVersion(h.opts.Parameters.Override.HelmRepoURL, component.Name())
			if err != nil {
				return errors.Wrapf(err, "while getting latest version for %s", component.Name())
			}
		}

		status.Step("Installing %s Helm Chart", component.Name())
		err := component.InstallUpgrade(version)
		if err != nil {
			return err
		}
	}
	return nil
}

func shouldSkipTheComponent(name string, skipList []string) bool {
	for _, skip := range skipList {
		if skip == name {
			return true
		}
	}
	return false
}

func (h *Helm) InstallCRD() error {
	resp, err := http.Get(CRDUrl)
	if err != nil {
		return errors.Wrapf(err, "while getting CRD %s", CRDUrl)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "while downloading CRD %s", CRDUrl)
	}
	return createObject(h.configuration, content)
}

func createObject(configuration *action.Configuration, content []byte) error {
	res, err := configuration.KubeClient.Build(bytes.NewBuffer(content), true)
	if err != nil {
		return errors.Wrap(err, "while validating the object")
	}

	if _, err := configuration.KubeClient.Create(res); err != nil {
		// If the error is CRD already exists, return.
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return errors.Wrapf(err, "while creating the object")
	}
	return nil
}
