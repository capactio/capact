// Package manifestgen_test is based on golden file pattern.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// To update golden files, run:
//   go test ./internal/cli/manifestgen/... -test.update-golden
package manifestgen_test

import (
	"fmt"
	"testing"

	"capact.io/capact/cmd/cli/cmd/manifest/generate/common"
	"capact.io/capact/internal/cli/manifestgen"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
)

func TestGenerateAttributeManifests(t *testing.T) {
	cfg := &manifestgen.AttributeConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.attribute.group.test",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateAttributeTemplatingConfig(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateInputTypeManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.type.input.test",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateInputTypeTemplatingConfig(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateOutputTypeManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.type.output.test",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateOutputTypeTemplatingConfig(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateTypeManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.type.test.empty",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateTypeTemplatingConfig(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateInterfaceGroupManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceGroupConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.interface.grouptest",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateInterfaceGroupTemplatingConfig(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateInterfaceManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     "cap.interface.group.test",
				Revision: "0.2.0",
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: ptr.String("https://example.com"),
			SupportURL:       ptr.String("https://example.com"),
			IconURL:          ptr.String("https://example.com/icon.png"),
			Maintainers: []types.Maintainer{
				{
					Email: "dev@example.com",
					Name:  ptr.String("Example Dev"),
					URL:   ptr.String("https://example.com"),
				},
			},
		},
	}

	manifests, err := manifestgen.GenerateInterfaceManifests(cfg)
	require.NoError(t, err)

	for name, manifestData := range manifests {
		filename := fmt.Sprintf("%s.yaml", name)
		golden.Assert(t, string(manifestData), filename)
	}
}

func TestGenerateEmptyImplementationManifests(t *testing.T) {
	tests := []struct {
		name string
		cfg  *manifestgen.EmptyImplementationConfig
	}{
		{
			name: "Implementation manifests",
			cfg: &manifestgen.EmptyImplementationConfig{
				AdditionalInputTypeName: "additional-parameters",
				ImplementationConfig: manifestgen.ImplementationConfig{
					Config: manifestgen.Config{
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.empty.test",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
					},
					InterfacePathWithRevision: "cap.interface.group.test:0.2.0",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifests, err := manifestgen.GenerateEmptyManifests(test.cfg)
			require.NoError(t, err)

			for name, manifestData := range manifests {
				filename := fmt.Sprintf("%s.yaml", name)
				golden.Assert(t, string(manifestData), filename)
			}
		})
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
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.terraform.generic.test",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
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
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.terraform.aws.test",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
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
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.terraform.gcp.test",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
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
				golden.Assert(t, string(manifestData), filename)
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
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.helm.test",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
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
						ManifestRef: types.ManifestRef{
							Path:     "cap.implementation.helm.test-generated-schema",
							Revision: "0.1.0",
						},
					},
					Metadata: types.ImplementationMetadata{
						DocumentationURL: ptr.String("https://example.com"),
						SupportURL:       ptr.String("https://example.com"),
						Maintainers: []types.Maintainer{
							{
								Email: "dev@example.com",
								Name:  ptr.String("Example Dev"),
								URL:   ptr.String("https://example.com"),
							},
						},
						License: types.License{
							Name: common.ApacheLicense,
						},
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
				golden.Assert(t, string(manifestData), filename)
			}
		})
	}
}
