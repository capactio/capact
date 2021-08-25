package capact

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/maps"
	"capact.io/capact/internal/ptr"

	util "github.com/Masterminds/goutils"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/strvals"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/strings/slices"
)

// Component is a Capact component which can be installed in the environement
type Component interface {
	InstallUpgrade(ctx context.Context, version string) (*release.Release, error)
	Name() string
	Chart() string
	withOptions(*Options)
	withConfiguration(*action.Configuration)
	withWriter(io.Writer)
}

type components []Component

func (i components) All() []string {
	var all []string
	for _, c := range i {
		all = append(all, c.Name())
	}
	return all
}

// ComponentData information about component
type ComponentData struct {
	ReleaseName string
	LocalPath   string
	ChartName   string
	Wait        bool

	Resources *Resources
	Overrides map[string]interface{}

	configuration *action.Configuration
	opts          *Options

	writer io.Writer
}

// Name of the Release
func (c *ComponentData) Name() string {
	return c.ReleaseName
}

// Chart name of the component
func (c *ComponentData) Chart() string {
	if c.ChartName != "" {
		return c.ChartName
	}
	return c.ReleaseName
}

func (c *ComponentData) installAction(version string) *action.Install {
	installCli := action.NewInstall(c.configuration)

	installCli.DryRun = c.opts.DryRun
	installCli.Namespace = c.opts.Namespace
	installCli.Timeout = c.opts.Timeout

	installCli.ChartPathOptions.Version = version
	installCli.ChartPathOptions.RepoURL = c.opts.Parameters.Override.HelmRepoURL

	installCli.NameTemplate = c.Name()
	installCli.ReleaseName = c.Name()

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

func (c *ComponentData) runUpgrade(upgradeCli *action.Upgrade, values map[string]interface{}) (*release.Release, error) {
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
		location = c.Chart()
	}

	chartPath, err = upgradeCli.ChartPathOptions.LocateChart(location, &cli.EnvSettings{
		RepositoryCache: RepositoryCache,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while locating Helm chart")
	}

	chartData, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "while loading Helm chart")
	}

	r, err := upgradeCli.Run(c.Name(), chartData, values)
	if err != nil {
		return nil, errors.Wrapf(err, "while upgrading Helm chart [%s]", c.Name())
	}
	return r, nil
}

func (c *ComponentData) runInstall(installCli *action.Install, values map[string]interface{}) (*release.Release, error) {
	var chartPath string
	var err error
	var location string

	if installCli.Version == LocalVersionTag {
		location = c.LocalPath
	} else {
		location = c.Chart()
	}

	chartPath, err = installCli.ChartPathOptions.LocateChart(location, &cli.EnvSettings{
		RepositoryCache: RepositoryCache,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while locating Helm chart")
	}

	chartData, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "while loading Helm chart")
	}

	r, err := installCli.Run(chartData, values)
	if err != nil {
		return nil, errors.Wrapf(err, "while installing Helm chart [%s]", installCli.ReleaseName)
	}

	return r, nil
}

// based on https://github.com/helm/helm/blob/433b90c4b6010415524bfd98b77efca0e6ec7205/cmd/helm/status.go#L112
func (h Helm) writeStatus(out io.Writer, r *release.Release) {
	if r == nil {
		return
	}
	fmt.Fprintf(out, "\tNAME: %s\n", r.Name)
	if !r.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "\tLAST DEPLOYED: %s\n", r.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Fprintf(out, "\tNAMESPACE: %s\n", r.Namespace)
	fmt.Fprintf(out, "\tSTATUS: %s\n", r.Info.Status.String())
	fmt.Fprintf(out, "\tREVISION: %d\n", r.Version)
	fmt.Fprintf(out, "\tDESCRIPTION: %s\n", r.Info.Description)
}

func (h Helm) writeHelmDetails(out io.Writer) {
	fmt.Fprintf(out, "\n  Installation details:\n")
	fmt.Fprintf(out, "\tVersion: %s\n", h.opts.Parameters.Version)

	helmRepo := h.opts.Parameters.Override.HelmRepoURL
	if h.opts.Parameters.Version == LocalVersionTag {
		helmRepo = LocalChartsPath
	}
	fmt.Fprintf(out, "\tHelm repository: %s\n\n", helmRepo)
}

// Components is a list of all Capact components available to install.
var Components = components{
	&Neo4j{
		ComponentData{
			ReleaseName: "neo4j",
			LocalPath:   path.Join(LocalChartsPath, "neo4j"),
			Wait:        true,
		},
	},
	&IngressController{
		ComponentData{
			ReleaseName: "ingress-nginx",
			ChartName:   "ingress-controller",
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

// Neo4j component
type Neo4j struct {
	ComponentData
}

// IngressController component
type IngressController struct {
	ComponentData
}

// Argo component
type Argo struct {
	ComponentData
}

// CertManager component
type CertManager struct {
	ComponentData
}

// Kubed component
type Kubed struct {
	ComponentData
}

// Monitoring component
type Monitoring struct {
	ComponentData
}

// Capact component
type Capact struct {
	ComponentData
}

// InstallUpgrade upgrades or if not available, installs the component
func (n *Neo4j) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	upgradeCli := n.upgradeAction(version)

	values := maps.Merge(n.opts.Parameters.Override.Neo4jValues.AsMap(), n.Overrides)

	return n.runUpgrade(upgradeCli, values)
}

// InstallUpgrade upgrades or if not available, installs the component
func (a *Argo) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	upgradeCli := a.upgradeAction(version)

	values := make(map[string]interface{})

	installed, err := a.isInstalled()
	if err != nil {
		return nil, errors.Wrap(err, "while getting Argo Helm release")
	}

	if !installed && !a.areMinioCredentialsProvided(a.Overrides) {
		accessKey, err := util.CryptoRandomAlphaNumeric(10)
		if err != nil {
			return nil, errors.Wrap(err, "while generating accessKey")
		}
		secretKey, err := util.CryptoRandomAlphaNumeric(40)
		if err != nil {
			return nil, errors.Wrap(err, "while generating secretKey")
		}

		credentials := map[string]interface{}{
			"minio": map[string]interface{}{
				"accessKey": map[string]interface{}{
					"password": accessKey,
				},
				"secretKey": map[string]interface{}{
					"password": secretKey,
				},
			},
		}

		values = maps.Merge(values, credentials)
	}

	values = maps.Merge(values, a.Overrides)

	return a.runUpgrade(upgradeCli, values)
}

func (a *Argo) isInstalled() (bool, error) {
	getAction := action.NewGet(a.configuration)
	_, err := getAction.Run(a.ReleaseName)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "while checking if the Argo release exists")
	}

	return true, nil
}

func (a *Argo) areMinioCredentialsProvided(values map[string]interface{}) bool {
	minio, ok := values["minio"].(map[string]interface{})
	if !ok {
		// if minio key is not set
		return false
	}

	if existingSecret, ok := minio["existingSecret"]; ok && existingSecret != "" {
		// if minio.existingSecret is set
		return true
	}

	accessKey, ok := minio["accessKey"].(map[string]interface{})
	if !ok {
		// if minio.accessKey is not set
		return false
	}

	if accessKeyPassword, ok := accessKey["password"]; !ok || accessKeyPassword == "" {
		// if minio.accessKey.password is not set
		return false
	}

	secretKey, ok := minio["secretKey"].(map[string]interface{})
	if !ok {
		// if minio.secretKey is not set
		return false
	}

	if secretKeyPassword, ok := secretKey["password"]; !ok || secretKeyPassword == "" {
		// if minio.secretKey.password is not set
		return false
	}

	// minio.accessKey.password and minio.secretKey.password is set
	return true
}

// InstallUpgrade upgrades or if not available, installs the component
func (i *IngressController) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	var err error
	upgradeCli := i.upgradeAction(version)

	values := map[string]interface{}{}

	switch i.opts.Environment {
	case KindEnv:
		values, err = ValuesFromString(ingressKindOverridesYaml)
		if err != nil {
			return nil, errors.Wrap(err, "while converting override values")
		}

	case EKSEnv:
		values, err = ValuesFromString(ingressEksOverridesYaml)
		if err != nil {
			return nil, errors.Wrap(err, "while converting override values")
		}
	}

	for _, value := range i.opts.Parameters.Override.IngressStringOverrides {
		if err := strvals.ParseInto(value, values); err != nil {
			return nil, errors.Wrap(err, "failed parsing passed overrides")
		}
	}
	return i.runUpgrade(upgradeCli, values)
}

// InstallUpgrade upgrades or if not available, installs the component
func (c *CertManager) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	var err error
	upgradeCli := c.upgradeAction(version)

	values := map[string]interface{}{}
	switch c.opts.Environment {
	case EKSEnv:
		values, err = ValuesFromString(certManagerEksOverridesYaml)
		if err != nil {
			return nil, errors.Wrap(err, "while converting override values")
		}
	}

	for _, value := range c.opts.Parameters.Override.CertManagerStringOverrides {
		if err := strvals.ParseInto(value, values); err != nil {
			return nil, errors.Wrap(err, "failed parsing passed overrides")
		}
	}

	r, err := c.runUpgrade(upgradeCli, values)
	if err != nil {
		return nil, errors.Wrap(err, "while installing cert-manager")
	}

	if c.opts.Environment != KindEnv {
		return nil, nil
	}

	restConfig, err := c.configuration.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while getting k8s REST config")
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
	err = CreateUpdateSecret(ctx, restConfig, secret, c.opts.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s Secret", certManagerSecretName)
	}

	// Not using cert-manager types as it's conflicting with argo deps
	issuer := fmt.Sprintf(issuerTemplate, clusterIssuerName, certManagerSecretName)
	err = createObject(c.configuration, []byte(issuer))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s ClusterIssuer", clusterIssuerName)
	}
	return r, nil
}

// InstallUpgrade upgrades or if not available, installs the component
func (k *Kubed) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	restConfig, err := k.configuration.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while getting k8s REST config")
	}

	upgradeCli := k.upgradeAction(version)
	values := map[string]interface{}{}
	r, err := k.runUpgrade(upgradeCli, values)
	if err != nil {
		return nil, errors.Wrap(err, "while running action")
	}

	err = AnnotateSecret(ctx, restConfig, "argo-minio", k.opts.Namespace, "kubed.appscode.com/sync", "")
	return r, errors.Wrap(err, "while annotating secret")
}

// InstallUpgrade upgrades or if not available, installs the component
func (c *Capact) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	upgradeCli := c.upgradeAction(version)

	capactValues := c.opts.Parameters.Override.CapactValues.AsMap()

	switch c.opts.Environment {
	case KindEnv:
		values, err := ValuesFromString(capactKindOverridesYaml)
		if err != nil {
			return nil, errors.Wrap(err, "while converting override values")
		}
		capactValues = maps.Merge(values, capactValues)
	}

	for _, value := range c.opts.Parameters.Override.CapactStringOverrides {
		if err := strvals.ParseInto(value, capactValues); err != nil {
			return nil, errors.Wrap(err, "failed parsing passed overrides")
		}
	}

	return c.runUpgrade(upgradeCli, capactValues)
}

// InstallUpgrade upgrades or if not available, installs the component
func (m *Monitoring) InstallUpgrade(ctx context.Context, version string) (*release.Release, error) {
	upgradeAction := m.upgradeAction(version)

	values := map[string]interface{}{}
	return m.runUpgrade(upgradeAction, values)
}

// GetActionConfiguration gets Helm action.Configuration
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

// Helm installs Helm components
type Helm struct {
	configuration *action.Configuration
	opts          Options
}

// NewHelm creates a Helm struct
func NewHelm(configuration *action.Configuration, opts Options) *Helm {
	if opts.Parameters.IncreaseResourceLimits {
		opts.Parameters.Override.CapactValues.Gateway.Resources = IncreasedGatewayResources()
		opts.Parameters.Override.CapactValues.HubPublic.Resources = IncreasedHubPublicResources()
		opts.Parameters.Override.CapactValues.HubLocal.Resources = IncreasedHubLocalResources()
		opts.Parameters.Override.Neo4jValues.Neo4j.Core.Resources = IncreasedNeo4jResources()
	}
	return &Helm{configuration: configuration, opts: opts}
}

// InstallComponents installs Helm components
func (h *Helm) InstallComponents(ctx context.Context, w io.Writer, status *printer.Status) error {
	if h.opts.Verbose {
		status.Step("Resolving installation config")
		status.End(true)
		h.writeHelmDetails(w)
	}

	for _, component := range Components {
		if !slices.Contains(h.opts.InstallComponents, component.Name()) {
			continue
		}

		component.withOptions(&h.opts)
		component.withConfiguration(h.configuration)
		component.withWriter(w)

		status.Step("Installing %s Helm chart", component.Name())
		newRelease, err := component.InstallUpgrade(ctx, h.opts.Parameters.Version)
		status.End(err == nil)
		if err != nil {
			return err
		}
		if h.opts.Verbose {
			h.writeStatus(w, newRelease)
		}
	}
	return nil
}

// InstallCRD installs Capact CRD
func (h *Helm) InstallCRD() error {
	var reader io.Reader
	if h.opts.Parameters.Version == LocalVersionTag {
		f, err := os.Open(CRDLocalPath)
		if err != nil {
			return errors.Wrapf(err, "while opening local CRD file%s", CRDLocalPath)
		}
		defer f.Close()
		reader = f
	} else {
		resp, err := http.Get(CRDUrl)
		if err != nil {
			return errors.Wrapf(err, "while getting CRD %s", CRDUrl)
		}
		defer resp.Body.Close()
		reader = resp.Body
	}

	content, err := ioutil.ReadAll(reader)
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

	// 5 exponential retries: ~2µs ~12ms ~70ms ~325ms ~1.63s
	retryBackoff := retry.DefaultBackoff
	retryBackoff.Steps = 5
	err = retry.OnError(retryBackoff, differentThanAlreadyExistErr, func() error {
		_, err = configuration.KubeClient.Create(res)
		// You have to return err itself here (not wrapped inside another error)
		// so we can identify it correctly.
		return err
	})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		// May be conflict if max retries were hit, or may be something unrelated like permissions error
		return err
	}

	return nil
}

func differentThanAlreadyExistErr(err error) bool {
	return !apierrors.IsAlreadyExists(err)
}
