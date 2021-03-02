package client

import (
	"context"
	"fmt"
	"sync"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"

	"github.com/pkg/errors"

	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"

	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type OCHClient interface {
	GetInterfaceLatestRevisionString(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (string, error)
	GetImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error)
	ListTypeInstancesTypeRef(ctx context.Context) ([]ochlocalgraphql.TypeReference, error)
	GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error)
}

type PolicyEnforcedClient struct {
	ochCli OCHClient
	policy clusterpolicy.ClusterPolicy
	mu     sync.RWMutex
}

func NewPolicyEnforcedClient(ochCli OCHClient) *PolicyEnforcedClient {
	return &PolicyEnforcedClient{ochCli: ochCli}
}

func (e *PolicyEnforcedClient) ListImplementationRevisionForInterface(ctx context.Context, interfaceRef ochpublicgraphql.InterfaceReference) ([]ochpublicgraphql.ImplementationRevision, clusterpolicy.Rule, error) {
	if interfaceRef.Revision == "" {
		interfaceRevision, err := e.ochCli.GetInterfaceLatestRevisionString(ctx, interfaceRef)
		if err != nil {
			return nil, clusterpolicy.Rule{}, errors.Wrap(err, "while fetching latest Interface revision string")
		}

		interfaceRef.Revision = interfaceRevision
	}

	rules := e.findRulesForInterface(interfaceRef)
	if len(rules.OneOf) == 0 {
		return nil, clusterpolicy.Rule{}, nil
	}

	typeInstanceValues, err := e.listCurrentTypeInstanceValues(ctx)
	if err != nil {
		return nil, clusterpolicy.Rule{}, err
	}
	typeInstanceValues = append(typeInstanceValues, e.constantTypeInstanceValues()...)

	implementations, rule, err := e.findImplementationsForRules(ctx, interfaceRef, rules, typeInstanceValues)
	if err != nil {
		return nil, clusterpolicy.Rule{}, err
	}

	return implementations, rule, nil
}

func (e *PolicyEnforcedClient) ListTypeInstancesToInjectBasedOnPolicy(policyRule clusterpolicy.Rule, implRev ochpublicgraphql.ImplementationRevision) []types.InputTypeInstanceRef {
	if len(policyRule.InjectTypeInstances) == 0 {
		return nil
	}

	var typeInstancesToInject []types.InputTypeInstanceRef
	for _, typeInstance := range policyRule.InjectTypeInstances {
		alias, found := e.findAliasForTypeInstance(typeInstance, implRev)
		if !found {
			// Implementation doesn't require such TypeInstance, skip injecting it
			continue
		}

		typeInstanceToInject := types.InputTypeInstanceRef{
			Name: alias,
			ID:   typeInstance.ID,
		}

		typeInstancesToInject = append(typeInstancesToInject, typeInstanceToInject)
	}

	return typeInstancesToInject
}

func (e *PolicyEnforcedClient) GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error) {
	return e.ochCli.GetInterfaceRevision(ctx, ref)
}

// SetPolicy sets policy to use. This setter is thread safe.
func (e *PolicyEnforcedClient) SetPolicy(policy clusterpolicy.ClusterPolicy) {
	e.mu.Lock()
	e.policy = policy
	e.mu.Unlock()
}

// Policy gets policy which the Client uses. This getter is thread safe.
func (e *PolicyEnforcedClient) Policy() clusterpolicy.ClusterPolicy {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.policy
}

func (e *PolicyEnforcedClient) findRulesForInterface(interfaceRef ochpublicgraphql.InterfaceReference) clusterpolicy.Rules {
	ruleKeysToCheck := []clusterpolicy.InterfacePath{
		clusterpolicy.InterfacePath(fmt.Sprintf("%s:%s", interfaceRef.Path, interfaceRef.Revision)),
		clusterpolicy.InterfacePath(interfaceRef.Path),
		clusterpolicy.AnyInterfacePath,
	}

	for _, ruleKey := range ruleKeysToCheck {
		rules, exists := e.Policy().Rules[ruleKey]
		if !exists {
			continue
		}

		return rules
	}

	return clusterpolicy.Rules{}
}

func (e *PolicyEnforcedClient) findImplementationsForRules(
	ctx context.Context,
	interfaceRef ochpublicgraphql.InterfaceReference,
	rules clusterpolicy.Rules,
	currentTypeInstances []*ochpublicgraphql.TypeInstanceValue,
) ([]ochpublicgraphql.ImplementationRevision, clusterpolicy.Rule, error) {
	for _, rule := range rules.OneOf {
		filter := e.implementationConstraintsToOCHFilter(rule.ImplementationConstraints)
		filter.RequirementsSatisfiedBy = currentTypeInstances

		implementations, err := e.ochCli.GetImplementationRevisionsForInterface(
			ctx,
			interfaceRef,
			public.WithFilter(filter),
			public.WithSortingByPathAscAndRevisionDesc(),
		)
		switch err := errors.Cause(err).(type) {
		case nil:
		case *public.ImplementationRevisionNotFoundError:
			continue
		default:
			return nil, clusterpolicy.Rule{}, err
		}

		if len(implementations) == 0 {
			continue
		}

		return implementations, rule, nil
	}

	return nil, clusterpolicy.Rule{}, nil
}

func (e *PolicyEnforcedClient) findAliasForTypeInstance(typeInstance clusterpolicy.TypeInstanceToInject, implRev ochpublicgraphql.ImplementationRevision) (string, bool) {
	if implRev.Spec == nil || len(implRev.Spec.Requires) == 0 {
		return "", false
	}

	for _, req := range implRev.Spec.Requires {
		var itemsToCheck []*ochpublicgraphql.ImplementationRequirementItem
		itemsToCheck = append(itemsToCheck, req.OneOf...)
		itemsToCheck = append(itemsToCheck, req.AllOf...)
		itemsToCheck = append(itemsToCheck, req.AnyOf...)

		for _, req := range itemsToCheck {
			if !e.isTypeRefValidAndEqual(typeInstance, req) {
				continue
			}

			return *req.Alias, true
		}
	}

	return "", false
}

func (e *PolicyEnforcedClient) implementationConstraintsToOCHFilter(constraints clusterpolicy.ImplementationConstraints) ochpublicgraphql.ImplementationRevisionFilter {
	filter := ochpublicgraphql.ImplementationRevisionFilter{}

	// Path
	if constraints.Path != nil {
		filter.PathPattern = constraints.Path
	}

	// Requires
	if constraints.Requires != nil && len(*constraints.Requires) > 0 {
		for _, reqConstraint := range *constraints.Requires {
			filter.Requires = append(filter.Requires, &ochpublicgraphql.TypeReferenceWithOptionalRevision{
				Path:     reqConstraint.Path,
				Revision: reqConstraint.Revision,
			})
		}
	}

	// Attributes
	if constraints.Attributes != nil && len(*constraints.Attributes) > 0 {
		for _, attrConstraint := range *constraints.Attributes {
			attrFilterRule := ochpublicgraphql.FilterRuleInclude
			filter.Attributes = append(filter.Attributes, &ochpublicgraphql.AttributeFilterInput{
				Path:     attrConstraint.Path,
				Rule:     &attrFilterRule,
				Revision: attrConstraint.Revision,
			})
		}
	}

	return filter
}

func (e *PolicyEnforcedClient) isTypeRefValidAndEqual(typeInstance clusterpolicy.TypeInstanceToInject, reqItem *ochpublicgraphql.ImplementationRequirementItem) bool {
	// check requirement item valid
	if reqItem == nil || reqItem.TypeRef == nil || reqItem.Alias == nil {
		return false
	}

	// check path
	if typeInstance.TypeRef.Path != reqItem.TypeRef.Path {
		return false
	}

	// check revision (if provided)
	if typeInstance.TypeRef.Revision != nil && *typeInstance.TypeRef.Revision != reqItem.TypeRef.Revision {
		return false
	}

	return true
}

func (e *PolicyEnforcedClient) listCurrentTypeInstanceValues(ctx context.Context) ([]*ochpublicgraphql.TypeInstanceValue, error) {
	currentTypeInstancesTypeRef, err := e.ochCli.ListTypeInstancesTypeRef(ctx)
	if err != nil {
		return nil, err
	}

	typeInstanceValues := e.typeInstancesToTypeInstanceValues(currentTypeInstancesTypeRef)

	return typeInstanceValues, nil
}

func (e *PolicyEnforcedClient) typeInstancesToTypeInstanceValues(in []ochlocalgraphql.TypeReference) []*ochpublicgraphql.TypeInstanceValue {
	var out []*ochpublicgraphql.TypeInstanceValue

	for _, typeRef := range in {
		out = append(out, &ochpublicgraphql.TypeInstanceValue{
			TypeRef: &ochpublicgraphql.TypeReferenceWithOptionalRevision{
				Path:     typeRef.Path,
				Revision: &typeRef.Revision,
			},
		})
	}

	return out
}

// TODO: Remove it as a part of TypeInstance autodiscovery when Engine starts
func (e *PolicyEnforcedClient) constantTypeInstanceValues() []*ochpublicgraphql.TypeInstanceValue {
	return []*ochpublicgraphql.TypeInstanceValue{
		{
			TypeRef: &ochpublicgraphql.TypeReferenceWithOptionalRevision{
				Path: "cap.core.type.platform.kubernetes",
			},
		},
	}
}
