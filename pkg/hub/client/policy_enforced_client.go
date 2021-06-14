package client

import (
	"context"
	"fmt"
	"sync"

	"capact.io/capact/pkg/engine/k8s/policy"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/pkg/errors"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

type HubClient interface {
	GetInterfaceLatestRevisionString(ctx context.Context, ref hubpublicgraphql.InterfaceReference) (string, error)
	ListImplementationRevisionsForInterface(ctx context.Context, ref hubpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]hubpublicgraphql.ImplementationRevision, error)
	ListTypeInstancesTypeRef(ctx context.Context) ([]hublocalgraphql.TypeInstanceTypeReference, error)
	FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference) (*hubpublicgraphql.InterfaceRevision, error)
}

type PolicyEnforcedClient struct {
	hubCli HubClient
	policy policy.Policy
	mu     sync.RWMutex
}

func NewPolicyEnforcedClient(hubCli HubClient) *PolicyEnforcedClient {
	return &PolicyEnforcedClient{hubCli: hubCli}
}

func (e *PolicyEnforcedClient) ListImplementationRevisionForInterface(ctx context.Context, interfaceRef hubpublicgraphql.InterfaceReference) ([]hubpublicgraphql.ImplementationRevision, policy.Rule, error) {
	if interfaceRef.Revision == "" {
		interfaceRevision, err := e.hubCli.GetInterfaceLatestRevisionString(ctx, interfaceRef)
		if err != nil {
			return nil, policy.Rule{}, errors.Wrap(err, "while fetching latest Interface revision string")
		}

		interfaceRef.Revision = interfaceRevision
	}

	rules := e.findRulesForInterface(interfaceRef)
	if len(rules.OneOf) == 0 {
		return nil, policy.Rule{}, nil
	}

	typeInstanceValues, err := e.listCurrentTypeInstanceValues(ctx)
	if err != nil {
		return nil, policy.Rule{}, err
	}
	typeInstanceValues = append(typeInstanceValues, e.constantTypeInstanceValues()...)

	implementations, rule, err := e.findImplementationsForRules(ctx, interfaceRef, rules, typeInstanceValues)
	if err != nil {
		return nil, policy.Rule{}, err
	}

	return implementations, rule, nil
}

func (e *PolicyEnforcedClient) ListTypeInstancesToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicgraphql.ImplementationRevision) []types.InputTypeInstanceRef {
	if policyRule.Inject == nil || len(policyRule.Inject.TypeInstances) == 0 {
		return nil
	}

	var typeInstancesToInject []types.InputTypeInstanceRef
	for _, typeInstance := range policyRule.Inject.TypeInstances {
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

func (e *PolicyEnforcedClient) FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference) (*hubpublicgraphql.InterfaceRevision, error) {
	return e.hubCli.FindInterfaceRevision(ctx, ref)
}

// SetPolicy sets policy to use. This setter is thread safe.
func (e *PolicyEnforcedClient) SetPolicy(policy policy.Policy) {
	e.mu.Lock()
	e.policy = policy
	e.mu.Unlock()
}

// Policy gets policy which the Client uses. This getter is thread safe.
func (e *PolicyEnforcedClient) Policy() policy.Policy {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.policy
}

func (e *PolicyEnforcedClient) findRulesForInterface(interfaceRef hubpublicgraphql.InterfaceReference) policy.RulesForInterface {
	rulesMap := e.rulesMapForPolicy(e.Policy())

	ruleKeysToCheck := []string{
		fmt.Sprintf("%s:%s", interfaceRef.Path, interfaceRef.Revision),
		interfaceRef.Path,
		policy.AnyInterfacePath,
	}

	for _, ruleKey := range ruleKeysToCheck {
		rules, exists := rulesMap[ruleKey]
		if !exists {
			continue
		}

		return rules
	}

	return policy.RulesForInterface{}
}

func (e *PolicyEnforcedClient) findImplementationsForRules(
	ctx context.Context,
	interfaceRef hubpublicgraphql.InterfaceReference,
	rules policy.RulesForInterface,
	currentTypeInstances []*hubpublicgraphql.TypeInstanceValue,
) ([]hubpublicgraphql.ImplementationRevision, policy.Rule, error) {
	for _, rule := range rules.OneOf {
		filter := e.implementationConstraintsToHubFilter(rule.ImplementationConstraints)
		filter.RequirementsSatisfiedBy = currentTypeInstances

		implementations, err := e.hubCli.ListImplementationRevisionsForInterface(
			ctx,
			interfaceRef,
			public.WithFilter(filter),
			public.WithSortingByPathAscAndRevisionDesc(),
		)
		if err != nil {
			return nil, policy.Rule{}, err
		}

		if len(implementations) == 0 {
			continue
		}

		return implementations, rule, nil
	}

	return nil, policy.Rule{}, nil
}

func (e *PolicyEnforcedClient) findAliasForTypeInstance(typeInstance policy.TypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) (string, bool) {
	if implRev.Spec == nil || len(implRev.Spec.Requires) == 0 {
		return "", false
	}

	for _, req := range implRev.Spec.Requires {
		var itemsToCheck []*hubpublicgraphql.ImplementationRequirementItem
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

func (e *PolicyEnforcedClient) implementationConstraintsToHubFilter(constraints policy.ImplementationConstraints) hubpublicgraphql.ImplementationRevisionFilter {
	filter := hubpublicgraphql.ImplementationRevisionFilter{}

	// Path
	if constraints.Path != nil {
		filter.PathPattern = constraints.Path
	}

	// Requires
	if constraints.Requires != nil && len(*constraints.Requires) > 0 {
		for _, reqConstraint := range *constraints.Requires {
			filter.Requires = append(filter.Requires, &hubpublicgraphql.TypeReferenceWithOptionalRevision{
				Path:     reqConstraint.Path,
				Revision: reqConstraint.Revision,
			})
		}
	}

	// Attributes
	if constraints.Attributes != nil && len(*constraints.Attributes) > 0 {
		for _, attrConstraint := range *constraints.Attributes {
			attrFilterRule := hubpublicgraphql.FilterRuleInclude
			filter.Attributes = append(filter.Attributes, &hubpublicgraphql.AttributeFilterInput{
				Path:     attrConstraint.Path,
				Rule:     &attrFilterRule,
				Revision: attrConstraint.Revision,
			})
		}
	}

	return filter
}

func (e *PolicyEnforcedClient) isTypeRefValidAndEqual(typeInstance policy.TypeInstanceToInject, reqItem *hubpublicgraphql.ImplementationRequirementItem) bool {
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

func (e *PolicyEnforcedClient) listCurrentTypeInstanceValues(ctx context.Context) ([]*hubpublicgraphql.TypeInstanceValue, error) {
	currentTypeInstancesTypeRef, err := e.hubCli.ListTypeInstancesTypeRef(ctx)
	if err != nil {
		return nil, err
	}

	typeInstanceValues := e.typeInstancesToTypeInstanceValues(currentTypeInstancesTypeRef)

	return typeInstanceValues, nil
}

func (e *PolicyEnforcedClient) typeInstancesToTypeInstanceValues(in []hublocalgraphql.TypeInstanceTypeReference) []*hubpublicgraphql.TypeInstanceValue {
	var out []*hubpublicgraphql.TypeInstanceValue

	for _, typeRef := range in {
		out = append(out, &hubpublicgraphql.TypeInstanceValue{
			TypeRef: &hubpublicgraphql.TypeReferenceWithOptionalRevision{
				Path:     typeRef.Path,
				Revision: &typeRef.Revision,
			},
		})
	}

	return out
}

// TODO: Remove it as a part of TypeInstance autodiscovery when Engine starts
func (e *PolicyEnforcedClient) constantTypeInstanceValues() []*hubpublicgraphql.TypeInstanceValue {
	return []*hubpublicgraphql.TypeInstanceValue{
		{
			TypeRef: &hubpublicgraphql.TypeReferenceWithOptionalRevision{
				Path: "cap.core.type.platform.kubernetes",
			},
		},
	}
}

func (e *PolicyEnforcedClient) rulesMapForPolicy(p policy.Policy) map[string]policy.RulesForInterface {
	rulesMap := map[string]policy.RulesForInterface{}
	for _, rule := range p.Rules {
		key := rule.Interface.Path
		if rule.Interface.Revision != nil {
			key = fmt.Sprintf("%s:%s", key, *rule.Interface.Revision)
		}
		rulesMap[key] = rule
	}

	return rulesMap
}
