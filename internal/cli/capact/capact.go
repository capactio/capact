package capact

import (
	"time"
)

const (
	// LatestVersionTag tag used to select latest version
	LatestVersionTag = "@latest"
	LocalVersionTag  = "@local"
	LocalDockerTag   = "dev"
	LocalDockerPath  = "local"
	KindEnv          = "kind"

	LocalChartsPath = "deploy/kubernetes/charts"
	HelmRepoLatest  = "https://storage.googleapis.com/capactio-latest-charts"
	HelmRepoStable  = "https://storage.googleapis.com/capactio-stable-charts"

	CRDUrl = "https://raw.githubusercontent.com/capactio/capact/main/deploy/kubernetes/crds/core.capact.io_actions.yaml"

	Name      = "capact"
	Namespace = "capact-system"

	RepositoryCache = "/tmp/helm"
)

// Options to set when interacting wit Capact
type Options struct {
	Name           string
	Namespace      string
	Environment    string
	SkipComponents []string
	SkipImages     []string
	FocusImages    []string
	DryRun         bool
	Timeout        time.Duration
	Parameters     InputParameters
}
