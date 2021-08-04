package content

import "fmt"

// Provider represents the provider type of the manifest
type Provider string

const (
	// ProviderAWS represents a AWS manifest type
	ProviderAWS Provider = "aws"

	// ProviderGCP respresents a GCP manifest type
	ProviderGCP Provider = "gcp"
)

// Set sets and validates the Provider from string
func (p *Provider) Set(s string) error {
	switch s {
	case string(ProviderAWS):
		*p = ProviderAWS
	case string(ProviderGCP):
		*p = ProviderGCP
	default:
		return fmt.Errorf(`provider not supported "%s"`, s)
	}

	return nil
}

func (p *Provider) String() string {
	return string(*p)
}

// Type returns the underlying type of the Provider
func (p *Provider) Type() string {
	return "string"
}
