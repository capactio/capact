package installation

import (
	"strings"

	"capact.io/capact/internal/logger"
	"github.com/vrischmann/envconfig"
)

// TypeInstancesConfig holds configuration for CapactRegister
type TypeInstancesConfig struct {
	Logger               logger.Config
	LocalHubEndpoint     string `envconfig:"default=http://capact-hub-local.capact-system/graphql"`
	HelmReleasesNSLookup LookupNS
	CapactReleaseName    string `envconfig:"default=capact"`
	HelmRepositoryPath   string `envconfig:"default=https://capactio-stable-charts.storage.googleapis.com"`
}

// LookupNS holds Namespaces names in which Capact components were installed.
type LookupNS map[string]struct{}

var _ envconfig.Unmarshaler = &LookupNS{}

// Unmarshal provides custom parsing for lookup namespaces syntax.
// Input is a comma separated list which is loaded into a map which provides O(1) for checking if a given element exists.
// Implements envconfig.Unmarshal interface.
func (m *LookupNS) Unmarshal(s string) error {
	in := strings.Split(s, ",")
	out := LookupNS{}
	for _, ns := range in {
		out[ns] = struct{}{}
	}
	*m = out

	return nil
}

// Contains returns true if a given ns was in LookupNS.
func (m LookupNS) Contains(ns string) bool {
	_, found := m[ns]
	return found
}
