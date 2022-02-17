package policy_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromYAMLBytes_ValidWithIgnoredTypeRef(t *testing.T) {
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

	expectedErrMessage := "while unmarshalling Policy from YAML bytes: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type policy.Policy"

	// when
	_, err := policy.FromYAMLString(in)

	// then
	require.Error(t, err)
	assert.EqualError(t, err, expectedErrMessage)
}

func loadInput(t *testing.T, path string) string {
	bytes, err := ioutil.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(bytes)
}

func fixValidPolicy() policy.Policy {
	return policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
				{
					Interface: types.ManifestRefWithOptRevision{
						Path:     "cap.interface.database.postgresql.install",
						Revision: ptr.String("0.1.0"),
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{
								Requires: &[]types.ManifestRefWithOptRevision{
									{
										Path:     "cap.type.gcp.auth.service-account",
										Revision: ptr.String("0.1.0"),
									},
								},
								Attributes: &[]types.ManifestRefWithOptRevision{
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
								RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "sample-uuid",
											Description: ptr.String("Google Cloud Platform Service Account"),
										},
									},
								},
								AdditionalParameters: []policy.AdditionalParametersToInject{
									{
										Name: "additional-parameters",
										Value: map[string]interface{}{
											"snapshot": true,
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "sample-uuid",
											Name: "sample-name",
										},
									},
								},
							},
						},
						{
							ImplementationConstraints: policy.ImplementationConstraints{
								Attributes: &[]types.ManifestRefWithOptRevision{
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
					Interface: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					OneOf: []policy.Rule{
						{
							ImplementationConstraints: policy.ImplementationConstraints{
								Requires: &[]types.ManifestRefWithOptRevision{
									{
										Path: "cap.core.type.platform.kubernetes",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
