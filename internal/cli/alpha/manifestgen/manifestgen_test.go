// Package manifestgen_test is based on golden file pattern.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// To update golden files, run:
//   go test ./internal/cli/alpha/manifestgen/... -test.update-golden
package manifestgen_test

import (
	"fmt"
	"testing"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
)

func TestGenerateAttributeManifests(t *testing.T) {
	cfg := &manifestgen.AttributeConfig{
		Config: manifestgen.Config{
			ManifestPath:     "cap.attribute.group.test",
			ManifestRevision: "0.2.0",
			ManifestMetadata: manifestgen.MetaDataInfo{
				DocumentationURL: "https://example.com",
				SupportURL:       "https://example.com",
				Maintainers: []common.Maintainers{
					{
						Email: "dev@example.com",
						Name:  "Example Dev",
						URL:   "https://example.com",
					},
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
func TestGenerateInterfaceManifests(t *testing.T) {
	cfg := &manifestgen.InterfaceConfig{
		InputPathWithRevision:  "cap.type.group.test-input:0.1.0",
		OutputPathWithRevision: "cap.type.group.config:0.1.0",
		Config: manifestgen.Config{
			ManifestPath:     "cap.interface.group.test",
			ManifestRevision: "0.2.0",
			ManifestMetadata: manifestgen.MetaDataInfo{
				DocumentationURL: "https://example.com",
				SupportURL:       "https://example.com",
				IconURL:          "https://example.com/icon.png",
				Maintainers: []common.Maintainers{
					{
						Email: "dev@example.com",
						Name:  "Example Dev",
						URL:   "https://example.com",
					},
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
						ManifestPath:     "cap.implementation.empty.test",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
						ManifestPath:     "cap.implementation.terraform.generic.test",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
						ManifestPath:     "cap.implementation.terraform.aws.test",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
						ManifestPath:     "cap.implementation.terraform.gcp.test",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
						ManifestPath:     "cap.implementation.helm.test",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
						ManifestPath:     "cap.implementation.helm.test-generated-schema",
						ManifestRevision: "0.1.0",
						ManifestMetadata: manifestgen.MetaDataInfo{
							DocumentationURL: "https://example.com",
							SupportURL:       "https://example.com",
							Maintainers: []common.Maintainers{
								{
									Email: "dev@example.com",
									Name:  "Example Dev",
									URL:   "https://example.com",
								},
							},
							License: types.License{
								Name: &common.ApacheLicense,
							},
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
