package clusterpolicy_test

import (
	"fmt"
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
	actual, err := clusterpolicy.FromYAMLString(in)

	// then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFromYAMLBytes_Invalid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/invalid.yaml")

	expectedConstraintErrs := []error{fmt.Errorf("2.0.0 does not have same major version as 0.1")}
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
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]clusterpolicy.TypeRef{
								{
									Path: "cap.core.type.platform.kubernetes",
								},
							},
						},
					},
				},
			},
			"cap.interface.database.postgresql.install:0.1.0": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]clusterpolicy.TypeRef{
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
								TypeRef: clusterpolicy.TypeRef{
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
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
		},
	}
}
