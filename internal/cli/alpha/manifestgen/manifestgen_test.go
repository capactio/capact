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
				Config: manifestgen.Config{
					ManifestPath:     "cap.implementation.terraform.generic.test",
					ManifestRevision: "0.1.0",
				},
				ModulePath:                "./testdata/terraform",
				ModuleSourceURL:           "https://example.com/module.tgz",
				InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
			},
		},
		{
			name: "Implementation manifests with AWS provider",
			cfg: &manifestgen.TerraformConfig{
				Config: manifestgen.Config{
					ManifestPath:     "cap.implementation.terraform.aws.test",
					ManifestRevision: "0.1.0",
				},
				ModulePath:                "./testdata/terraform",
				ModuleSourceURL:           "https://example.com/module.tgz",
				InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				Provider:                  manifestgen.ProviderAWS,
			},
		},
		{
			name: "Implementation manifests with GCP provider",
			cfg: &manifestgen.TerraformConfig{
				Config: manifestgen.Config{
					ManifestPath:     "cap.implementation.terraform.gcp.test",
					ManifestRevision: "0.1.0",
				},
				ModulePath:                "./testdata/terraform",
				ModuleSourceURL:           "https://example.com/module.tgz",
				InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				Provider:                  manifestgen.ProviderGCP,
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
	cfg := &manifestgen.HelmConfig{
		Config: manifestgen.Config{
			ManifestPath:     "cap.implementation.helm.test",
			ManifestRevision: "0.1.0",
		},
		ChartName:                 "postgresql",
		RepoURL:                   "https://charts.bitnami.com/bitnami",
		Version:                   "10.9.2",
		InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
	}

	manifests, err := manifestgen.GenerateHelmManifests(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, manifestData, filename)
	}
}
