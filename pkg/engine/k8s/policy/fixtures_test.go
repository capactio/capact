package policy_test

import (
	"context"
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

func fixPolicyWithoutTypeRef() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
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

func fixPolicyWithTypeRef() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.sample",
										Revision: "0.1.0",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.sample2",
										Revision: "0.2.0",
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

func fixComplexPolicyWithoutTypeRef() *policy.Policy {
	return &policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id1",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
									},
								},
							},
						},
					},
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id3",
									},
								},
							},
						},
					},
				},
			},
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.interface.productivity.mattermost.install",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id4",
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

func fixComplexPolicyWithTypeRef() *policy.Policy {
	return &policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id1",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.type1",
										Revision: "0.1.0",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.type2",
										Revision: "0.2.0",
									},
								},
							},
						},
					},
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id3",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.type3",
										Revision: "0.3.0",
									},
								},
							},
						},
					},
				},
			},
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.interface.productivity.mattermost.install",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id4",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.type4",
										Revision: "0.4.0",
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

type fakeHub struct {
	ShouldRun     bool
	ExpectedIDLen int
	IgnoreIDs     map[string]struct{}
}

func (f *fakeHub) FindTypeInstancesTypeRef(_ context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error) {
	if !f.ShouldRun {
		return nil, errors.New("shouldn't run")
	}
	if len(ids) != f.ExpectedIDLen {
		return nil, fmt.Errorf("invalid len: actual: %d, expected: %d", len(ids), f.ExpectedIDLen)
	}

	var idsToIncludeInResult []string
	for _, id := range ids {
		if _, ok := f.IgnoreIDs[id]; ok {
			continue
		}

		idsToIncludeInResult = append(idsToIncludeInResult, id)
	}

	res := make(map[string]gqllocalapi.TypeInstanceTypeReference)
	for _, id := range idsToIncludeInResult {
		idNumber := strings.TrimPrefix(id, "id")
		res[id] = gqllocalapi.TypeInstanceTypeReference{
			Path:     fmt.Sprintf("cap.type.type%s", idNumber),
			Revision: fmt.Sprintf("0.%s.0", idNumber),
		}
	}

	return res, nil
}
