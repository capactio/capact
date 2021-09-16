package capact

import (
	"time"

	"capact.io/capact/internal/multierror"

	"github.com/pkg/errors"
)

const (
	// LatestVersionTag tag used to select latest version
	LatestVersionTag = "@latest"
	// LocalVersionTag tag used to select local charts and images
	LocalVersionTag = "@local"
	// LocalDockerTag tag used when building local images
	LocalDockerTag = "dev"
	// LocalDockerPath path used when building local images
	LocalDockerPath = "local"

	// KindEnv default name for kind environment
	KindEnv = "kind"
	// GKEEnv default name for GKE environment
	GKEEnv = "gke"
	// EKSEnv default name for EKS environment
	EKSEnv = "eks"
	// K3dEnv default name for K3d environment
	K3dEnv = "k3d"

	// LocalChartsPath path to Helm charts in Capact repo
	LocalChartsPath = "deploy/kubernetes/charts"
	// HelmRepoLatest URL of the latest Capact charts repository
	HelmRepoLatest = "https://storage.googleapis.com/capactio-latest-charts"
	// HelmRepoStable URL of the stable Capact charts repository
	HelmRepoStable = "https://storage.googleapis.com/capactio-stable-charts"

	// CRDUrl Capact CRD URL
	CRDUrl = "https://raw.githubusercontent.com/capactio/capact/main/deploy/kubernetes/crds/core.capact.io_actions.yaml"
	// LocalCRDPath is a path to CRD definition in the repository
	LocalCRDPath = "deploy/kubernetes/crds/core.capact.io_actions.yaml"

	// Name Capact name
	Name = "capact"
	// Namespace Capact default namespace to install
	Namespace = "capact-system"

	// RepositoryCache Helm cache for repositories
	RepositoryCache = "/tmp/helm"

	// CertFile Capact Gateway certificate file name
	CertFile = "capact-local-ca.crt"
	// LinuxCertsPath path to Linux certificates directory
	LinuxCertsPath = "/usr/local/share/ca-certificates"
)

// Options to set when interacting wit Capact
type Options struct {
	Name                 string
	Namespace            string
	Environment          string
	InstallComponents    []string
	BuildImages          []string
	Timeout              time.Duration
	Parameters           InputParameters
	UpdateHostsFile      bool
	UpdateTrustedCerts   bool
	LocalRegistryEnabled bool

	// Helm client opts
	DryRun     bool
	ClientOnly bool
}

// Validate validates Capact install options.
func (o *Options) Validate() error {
	if err := o.Parameters.SetCapactValuesFromOverrides(); err != nil {
		return err
	}
	if o.LocalRegistryEnabled {
		return validateLocalRegistryOpts(o)
	}
	return nil
}

func validateLocalRegistryOpts(o *Options) error {
	merr := multierror.New()

	if o.Environment != K3dEnv {
		merr = multierror.Append(merr, errors.New("registry can be used only with k3d environment"))
	}
	if o.Parameters.Override.CapactValues.Global.ContainerRegistry.Path != "" {
		merr = multierror.Append(merr, errors.New("cannot specify container registry path override when local registry is enabled"))
	}

	return merr.ErrorOrNil()
}
