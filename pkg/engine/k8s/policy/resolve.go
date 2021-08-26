package policy

import (
	"context"

	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// HubClient defines Hub client which is able to find TypeInstance Type references.
type HubClient interface {
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]hublocalgraphql.TypeInstanceTypeReference, error)
}

// ResolveTypeInstanceMetadata resolves needed TypeInstance metadata based on IDs for a given Policy.
func ResolveTypeInstanceMetadata(ctx context.Context, hubCli HubClient, policy *Policy) error {
	if policy == nil {
		return errors.New("policy cannot be nil")
	}

	if hubCli == nil {
		return errors.New("hub client cannot be nil")
	}

	unresolvedReqTIs := policy.filterRequiredTypeInstances(filterReqTypeInstancesWithEmptyTypeRef)
	unresolvedAddTIs := policy.filterAdditionalTypeInstances(filterAddTypeInstancesWithEmptyTypeRef)

	var idsToQuery []string
	for _, ti := range unresolvedReqTIs {
		idsToQuery = append(idsToQuery, ti.ID)
	}
	for _, ti := range unresolvedAddTIs {
		idsToQuery = append(idsToQuery, ti.ID)
	}

	if len(idsToQuery) == 0 {
		return nil
	}

	res, err := hubCli.FindTypeInstancesTypeRef(ctx, idsToQuery)
	if err != nil {
		return errors.Wrap(err, "while finding TypeRef for TypeInstances")
	}

	resolveTypeRefsForRequiredTypeInstances(policy, res)
	resolveTypeRefsForAdditionalTypeInstances(policy, res)

	err = policy.ValidateTypeInstancesMetadata()
	if err != nil {
		return errors.Wrap(err, "while TypeInstance metadata validation after resolving TypeRefs")
	}

	return nil
}

// AreTypeInstancesMetadataResolved returns whether every TypeInstance has metadata resolved.
func (in *Policy) AreTypeInstancesMetadataResolved() bool {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterReqTypeInstancesWithEmptyTypeRef)
	return len(unresolvedTypeInstances) == 0
}

func resolveTypeRefsForRequiredTypeInstances(policy *Policy, typeRefs map[string]hublocalgraphql.TypeInstanceTypeReference) {
	for ruleIdx, rule := range policy.Rules {
		for ruleItemIdx, ruleItem := range rule.OneOf {
			if ruleItem.Inject == nil {
				continue
			}
			for reqTIIdx, reqTI := range ruleItem.Inject.RequiredTypeInstances {
				typeRef, exists := typeRefs[reqTI.ID]
				if !exists {
					continue
				}

				policy.Rules[ruleIdx].OneOf[ruleItemIdx].Inject.RequiredTypeInstances[reqTIIdx].TypeRef = &types.ManifestRef{
					Path:     typeRef.Path,
					Revision: typeRef.Revision,
				}
			}
		}
	}
}

func resolveTypeRefsForAdditionalTypeInstances(policy *Policy, typeRefs map[string]hublocalgraphql.TypeInstanceTypeReference) {
	for ruleIdx, rule := range policy.Rules {
		for ruleItemIdx, ruleItem := range rule.OneOf {
			if ruleItem.Inject == nil {
				continue
			}
			for reqTIIdx, reqTI := range ruleItem.Inject.AdditionalTypeInstances {
				typeRef, exists := typeRefs[reqTI.ID]
				if !exists {
					continue
				}

				policy.Rules[ruleIdx].OneOf[ruleItemIdx].Inject.AdditionalTypeInstances[reqTIIdx].TypeRef = &types.ManifestRef{
					Path:     typeRef.Path,
					Revision: typeRef.Revision,
				}
			}
		}
	}
}
