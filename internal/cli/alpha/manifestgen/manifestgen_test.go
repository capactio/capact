package manifestgen_test

import (
	"fmt"
	"testing"

	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
)

func TestGenerateInterfaceManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestPath:     "cap.interface.group.test",
			ManifestRevision: "0.2.0",
		},
	}

	manifests, err := manifestgen.GenerateInterfaceManifests(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, manifestData, filename)
	}
}

func TestGenerateTerraformImplementationManifests(t *testing.T) {
	tests := []struct {
		name string
		cfg  *manifestgen.TerraformConfig
	}{
		{
			name: "Implementation manifests",
			cfg: &manifestgen.TerraformConfig{
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestPath:     "cap.implementation.terraform.generic.test",
						ManifestRevision: "0.1.0",
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
				ModulePath:      "./testdata/terraform",
				ModuleSourceURL: "https://example.com/module.tgz",
			},
		},
		{
			name: "Implementation manifests with AWS provider",
			cfg: &manifestgen.TerraformConfig{
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestPath:     "cap.implementation.terraform.aws.test",
						ManifestRevision: "0.1.0",
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
				ModulePath:      "./testdata/terraform",
				ModuleSourceURL: "https://example.com/module.tgz",
				Provider:        manifestgen.ProviderAWS,
			},
		},
		{
			name: "Implementation manifests with GCP provider",
			cfg: &manifestgen.TerraformConfig{
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestPath:     "cap.implementation.terraform.gcp.test",
						ManifestRevision: "0.1.0",
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
				ModulePath:      "./testdata/terraform",
				ModuleSourceURL: "https://example.com/module.tgz",
				Provider:        manifestgen.ProviderGCP,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifests, err := manifestgen.GenerateTerraformManifests(test.cfg)
			require.NoError(t, err)

			for name, manifestData := range manifests {
				filename := fmt.Sprintf("%s.yaml", name)
				golden.Assert(t, manifestData, filename)
			}
		})
	}
}

func TestGenerateHelmImplementationManifests(t *testing.T) {
	tests := []struct {
		name string
		cfg  *manifestgen.HelmConfig
	}{
		{
			name: "Helm Implementation manifests with values.scheme.json",
			cfg: &manifestgen.HelmConfig{
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestPath:     "cap.implementation.helm.test",
						ManifestRevision: "0.1.0",
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
				ChartName:    "postgresql",
				ChartRepoURL: "https://charts.bitnami.com/bitnami",
				ChartVersion: "10.9.2",
			},
		},
		{
			name: "Helm Implementation manifests without values.scheme.json",
			cfg: &manifestgen.HelmConfig{
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestPath:     "cap.implementation.helm.test-generated-schema",
						ManifestRevision: "0.1.0",
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
				ChartName:    "dokuwiki",
				ChartRepoURL: "https://charts.bitnami.com/bitnami",
				ChartVersion: "11.2.3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifests, err := manifestgen.GenerateHelmManifests(test.cfg)
			require.NoError(t, err)

			for name, manifestData := range manifests {
				filename := fmt.Sprintf("%s.yaml", name)
				golden.Assert(t, manifestData, filename)
			}
		})
	}
}
