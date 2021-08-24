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

	err := resolveTypeRefsForRequiredTypeInstances(ctx, hubCli, policy)
	if err != nil {
		return err
	}

	err = policy.ValidateTypeInstancesMetadata()
	if err != nil {
		return errors.Wrap(err, "while TypeInstance metadata validation after resolving TypeRefs")
	}

	return nil
}

func resolveTypeRefsForRequiredTypeInstances(ctx context.Context, hubCli HubClient, policy *Policy) error {
	unresolvedTypeInstances := policy.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)

	var idsToQuery []string
	for _, ti := range unresolvedTypeInstances {
		idsToQuery = append(idsToQuery, ti.ID)
	}

	if len(idsToQuery) == 0 {
		return nil
	}

	res, err := hubCli.FindTypeInstancesTypeRef(ctx, idsToQuery)
	if err != nil {
		return errors.Wrap(err, "while finding TypeRef for TypeInstances")
	}

	for ruleIdx, rule := range policy.Rules {
		for ruleItemIdx, ruleItem := range rule.OneOf {
			if ruleItem.Inject == nil {
				continue
			}
			for reqTIIdx, reqTI := range ruleItem.Inject.RequiredTypeInstances {
				typeRef, exists := res[reqTI.ID]
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

	return nil
}
