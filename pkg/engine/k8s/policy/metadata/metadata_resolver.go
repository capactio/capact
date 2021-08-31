package metadata

import (
	"context"
	"fmt"

	"capact.io/capact/internal/multierror"

	"capact.io/capact/pkg/engine/k8s/policy"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	multierr "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// HubClient defines Hub client which is able to find TypeInstance Type references.
type HubClient interface {
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]hublocalgraphql.TypeInstanceTypeReference, error)
}

// Resolver resolves Policy metadata against Hub.
type Resolver struct {
	hubCli HubClient
}

// NewResolver returns new Resolver instance.
func NewResolver(hubCli HubClient) *Resolver {
	return &Resolver{hubCli: hubCli}
}

// ResolveTypeInstanceMetadata resolves needed TypeInstance metadata based on IDs for a given Policy.
func (r *Resolver) ResolveTypeInstanceMetadata(ctx context.Context, policy *policy.Policy) error {
	if policy == nil {
		return errors.New("policy cannot be nil")
	}

	if r.hubCli == nil {
		return errors.New("hub client cannot be nil")
	}

	unresolvedTIs := TypeInstanceIDsWithUnresolvedMetadataForPolicy(*policy)

	var idsToQuery []string
	for _, ti := range unresolvedTIs {
		idsToQuery = append(idsToQuery, ti.ID)
	}

	if len(idsToQuery) == 0 {
		return nil
	}

	res, err := r.hubCli.FindTypeInstancesTypeRef(ctx, idsToQuery)
	if err != nil {
		return errors.Wrap(err, "while finding TypeRef for TypeInstances")
	}

	if len(res) != len(idsToQuery) {
		multiErr := multierror.New()

		for _, ti := range unresolvedTIs {
			if typeRef, exists := res[ti.ID]; exists && typeRef.Path != "" && typeRef.Revision != "" {
				continue
			}

			multiErr = multierr.Append(multiErr, fmt.Errorf("missing Type reference for %s", ti.String(true)))
		}
		return multiErr
	}

	r.resolveTypeRefsForRequiredTypeInstances(policy, res)
	r.resolveTypeRefsForAdditionalTypeInstances(policy, res)

	return nil
}

func (r *Resolver) resolveTypeRefsForRequiredTypeInstances(policy *policy.Policy, typeRefs map[string]hublocalgraphql.TypeInstanceTypeReference) {
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

func (r *Resolver) resolveTypeRefsForAdditionalTypeInstances(policy *policy.Policy, typeRefs map[string]hublocalgraphql.TypeInstanceTypeReference) {
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
