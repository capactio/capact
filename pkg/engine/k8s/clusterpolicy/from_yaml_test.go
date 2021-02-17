package clusterpolicy_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

func TestFromYAMLBytes_Valid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/valid.yaml")
	expected := fixValidPolicy()

	// when
	actual, err := clusterpolicy.FromYAMLBytes(in)

	// then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFromYAMLBytes_Invalid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/invalid.yaml")
	expectedErr := clusterpolicy.NewUnsupportedAPIVersionError(clusterpolicy.SupportedAPIVersions.ToStringSlice())

	// when
	_, err := clusterpolicy.FromYAMLBytes(in)

	// then
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func loadInput(t *testing.T, path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return bytes
}

func fixValidPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.TypeRef{
								{
									Path: "cap.core.type.platform.kubernetes",
								},
							},
						},
					},
				},
			},
			"cap.interface.database.postgresql.install": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.TypeRef{
								{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.AttributeRef{
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
								TypeRef: types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Attributes: &[]types.AttributeRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Exact: &types.ImplementationRef{
								Path:     "cap.implementation.bitnami.postgresql.install",
								Revision: ptr.String("1.0.0"),
							},
						},
					},
				},
			},
		},
	}
}
