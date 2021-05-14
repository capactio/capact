package clusterpolicy_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromYAMLBytes_Valid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/valid.yaml")
	expected := fixValidPolicy()

	// when
	actual, err := clusterpolicy.FromYAMLString(in)

	// then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFromYAMLBytes_Invalid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/invalid.yaml")

	expectedConstraintErrs := []error{fmt.Errorf("2.0.0 does not have same major version as 0.2")}
	expectedErr := clusterpolicy.NewUnsupportedAPIVersionError(expectedConstraintErrs)

	// when
	_, err := clusterpolicy.FromYAMLString(in)

	// then
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func loadInput(t *testing.T, path string) string {
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return string(bytes)
}

func fixValidPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules: clusterpolicy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.ManifestRef{
								{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.ManifestRef{
								{
									Path:     "cap.attribute.cloud.provider.gcp",
									Revision: ptr.String("0.1.1"),
								},
								{
									Path:     "cap.core.attribute.workload.stateful",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
						InjectTypeInstances: []clusterpolicy.TypeInstanceToInject{
							{
								ID: "sample-uuid",
								TypeRef: types.ManifestRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Attributes: &[]types.ManifestRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.ManifestRef{
								{
									Path: "cap.core.type.platform.kubernetes",
								},
							},
						},
					},
				},
			},
		},
	}
}
