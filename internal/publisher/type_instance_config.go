package publisher

import "projectvoltron.dev/voltron/internal/logger"

// TypeInstancesConfig holds configuration for TypeInstances publisher
type TypeInstancesConfig struct {
	Logger               logger.Config
	LocalOCHEndpoint     string `envconfig:"default=http://voltron-och-local.voltron-system/graphql"`
	HelmReleasesNSLookup LookupNS
	VoltronReleaseName   string `envconfig:"default=voltron"`
	HelmRepositoryPath   string `envconfig:"default=https://capactio-awesome-charts.storage.googleapis.com"`
}

type LookupNS map[string]struct{}

// Unmarshal provides custom parsing for lookup namespaces syntax
// Implements envconfig.Unmarshal interface.
func (m *LookupNS) Unmarshal(in []string) error {
	out := LookupNS{}
	for _, ns := range in {
		out[ns] = struct{}{}
	}
	*m = out

	return nil
}

func (m LookupNS) Has(ns string) bool {
	_, found := m[ns]
	return found
}
