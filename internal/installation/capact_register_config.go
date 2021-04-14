package installation

import (
	"strings"

	"github.com/vrischmann/envconfig"
	"projectvoltron.dev/voltron/internal/logger"
)

// TypeInstancesConfig holds configuration for CapactRegister
type TypeInstancesConfig struct {
	Logger               logger.Config
	LocalOCHEndpoint     string `envconfig:"default=http://capact-och-local.capact-system/graphql"`
	HelmReleasesNSLookup LookupNS
	CapactReleaseName    string `envconfig:"default=capact"`
	// The `capactio-awesome-charts` bucket name is by design. We do not know if we will have to change our GCP project too,
	// so it will be easier to have the `capactio-charts` name not taken.
	HelmRepositoryPath string `envconfig:"default=https://capactio-awesome-charts.storage.googleapis.com"`
}

type LookupNS map[string]struct{}

var _ envconfig.Unmarshaler = &LookupNS{}

// Unmarshal provides custom parsing for lookup namespaces syntax.
// Input is a comma separated list which is loaded into a map which provides O(1) for checking if a given element exists.
// Implements envconfig.Unmarshal interface.]
func (m *LookupNS) Unmarshal(s string) error {
	in := strings.Split(s, ",")
	out := LookupNS{}
	for _, ns := range in {
		out[ns] = struct{}{}
	}
	*m = out

	return nil
}

func (m LookupNS) Contains(ns string) bool {
	_, found := m[ns]
	return found
}
