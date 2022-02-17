package metadata_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/pkg/errors"
)

func fixComplexPolicyWithoutTypeRef() *policy.Policy {
	return &policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id1",
										},
									},
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "id2",
											Description: ptr.String("ID 2"),
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id1",
											Name: "ID1",
										},
									},
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id3",
											Name: "ID3",
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id4",
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id5",
											Name: "ID5",
										},
									},
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id6",
											Name: "ID6",
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id7",
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id8",
											Name: "ID8",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		TypeInstance: policy.TypeInstancePolicy{
			Rules: []policy.RulesForTypeInstance{
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "id9",
							Description: ptr.String("ID 9"),
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.type.aws.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID: "id10",
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "id11",
							Description: ptr.String("ID 11"),
						},
					},
				},
			},
		},
	}
}

func fixComplexPolicyWithTypeRef() *policy.Policy {
	return &policy.Policy{
		Interface: policy.InterfacePolicy{
			Rules: policy.InterfaceRulesList{
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id1",
											TypeRef: &types.TypeRef{
												Path:     "cap.type.type1",
												Revision: "0.1.0",
											},
										},
									},
									{
										TypeInstanceReference: policy.TypeInstanceReference{
											ID:          "id2",
											Description: ptr.String("ID 2"),
											TypeRef: &types.TypeRef{
												Path:     "cap.type.type2",
												Revision: "0.2.0",
											},
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id1",
											Name: "ID1",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.type1",
											Revision: "0.1.0",
										},
									},
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id3",
											Name: "ID3",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.type3",
											Revision: "0.3.0",
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id4",
											TypeRef: &types.TypeRef{
												Path:     "cap.type.type4",
												Revision: "0.4.0",
											},
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id5",
											Name: "ID5",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.type5",
											Revision: "0.5.0",
										},
									},
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id6",
											Name: "ID6",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.type6",
											Revision: "0.6.0",
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
										TypeInstanceReference: policy.TypeInstanceReference{
											ID: "id7",
											TypeRef: &types.TypeRef{
												Path:     "cap.type.type7",
												Revision: "0.7.0",
											},
										},
									},
								},
								AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
									{
										AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
											ID:   "id8",
											Name: "ID8",
										},
										TypeRef: &types.ManifestRef{
											Path:     "cap.type.type8",
											Revision: "0.8.0",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		TypeInstance: policy.TypeInstancePolicy{
			Rules: []policy.RulesForTypeInstance{
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path:     "cap.type.aws.auth.credentials",
						Revision: ptr.String("0.1.0"),
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "id9",
							Description: ptr.String("ID 9"),
							TypeRef: &types.TypeRef{
								Path:     "cap.type.type9",
								Revision: "0.9.0",
							},
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.type.aws.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID: "id10",
							TypeRef: &types.TypeRef{
								Path:     "cap.type.type10",
								Revision: "0.10.0",
							},
						},
					},
				},
				{
					TypeRef: types.ManifestRefWithOptRevision{
						Path: "cap.*",
					},
					Backend: policy.TypeInstanceBackend{
						TypeInstanceReference: policy.TypeInstanceReference{
							ID:          "id11",
							Description: ptr.String("ID 11"),
							TypeRef: &types.TypeRef{
								Path:     "cap.type.type11",
								Revision: "0.11.0",
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

func (f *fakeHub) ListTypes(_ context.Context, opts ...public.TypeOption) ([]*hubpublicgraphql.Type, error) {
	if !f.ShouldRun {
		return nil, errors.New("shouldn't run")
	}

	var allTypes []*hubpublicgraphql.Type
	for id := 0; id < f.ExpectedIDLen; id++ {
		allTypes = append(allTypes, &hubpublicgraphql.Type{
			Path: fmt.Sprintf("cap.type.type%d", id),
			Revision: &hubpublicgraphql.TypeRevision{
				Revision: fmt.Sprintf("0.%d.0", id),
			},
		})
	}

	typeOpts := &public.TypeOptions{}
	typeOpts.Apply(opts...)

	if typeOpts.Filter.PathPattern == nil {
		return allTypes, nil
	}

	var out []*hubpublicgraphql.Type
	for _, item := range allTypes {
		matched, err := regexp.MatchString(*typeOpts.Filter.PathPattern, item.Path)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		out = append(out, item)
	}

	return out, nil
}
