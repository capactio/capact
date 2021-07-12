package policy_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromYAMLBytes_Valid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/valid.yaml")
	expected := fixValidPolicy()

	// when
	actual, err := policy.FromYAMLString(in)

	// then
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFromYAMLBytes_Invalid(t *testing.T) {
	// given
	in := loadInput(t, "testdata/invalid.yaml")

	expectedConstraintErrs := []error{fmt.Errorf("2.0.0 does not have same major version as 0.2")}
	expectedErr := policy.NewUnsupportedAPIVersionError(expectedConstraintErrs)

	// when
	_, err := policy.FromYAMLString(in)

	// then
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func loadInput(t *testing.T, path string) string {
	bytes, err := ioutil.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(bytes)
}

func fixValidPolicy() policy.Policy {
	return policy.Policy{
		APIVersion: policy.CurrentAPIVersion,
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRef{
					Path:     "cap.interface.database.postgresql.install",
					Revision: ptr.String("0.1.0"),
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
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
						Inject: &policy.InjectData{
							TypeInstances: []policy.TypeInstanceToInject{
								{
									ID: "sample-uuid",
									TypeRef: types.ManifestRef{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
							},
							AdditionalInput: map[string]interface{}{
								"snapshot": true,
							},
						},
					},
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Attributes: &[]types.ManifestRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
					{
						ImplementationConstraints: policy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
			{
				Interface: types.ManifestRef{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{
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
