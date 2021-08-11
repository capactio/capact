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

// HubClient interface aggregates methods for interacting with the Local and Public Hub.
type HubClient interface {
	GetInterfaceLatestRevisionString(ctx context.Context, ref hubpublicgraphql.InterfaceReference) (string, error)
	ListImplementationRevisionsForInterface(ctx context.Context, ref hubpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]hubpublicgraphql.ImplementationRevision, error)
	ListTypeInstancesTypeRef(ctx context.Context) ([]hublocalgraphql.TypeInstanceTypeReference, error)
	FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference, opts ...public.FindInterfaceRevisionOption) (*hubpublicgraphql.InterfaceRevision, error)
}

// PolicyEnforcedClient is a client, which can interact with the Local and Public Hub.
// It can be configured with policies to filter the Implementations returned by the Hub.
type PolicyEnforcedClient struct {
	hubCli               HubClient
	globalPolicy         policy.Policy
	actionPolicy         policy.Policy
	mergedPolicy         policy.Policy
	policyOrder          policy.MergeOrder
	workflowStepPolicies []policy.Policy
	mu                   sync.RWMutex
}

// NewPolicyEnforcedClient returns a new NewPolicyEnforcedClient.
func NewPolicyEnforcedClient(hubCli HubClient) *PolicyEnforcedClient {
	defaultOrder := policy.MergeOrder{policy.Action, policy.Global, policy.Workflow}
	return &PolicyEnforcedClient{hubCli: hubCli, policyOrder: defaultOrder}
}

// ListImplementationRevisionForInterface returns ImplementationRevisions
// for the given Interface and the current policy configuration.
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

// ListTypeInstancesToInjectBasedOnPolicy returns the input TypeInstance references,
// which have to be injected into the Action, based on the current policies.
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

// ListAdditionalInputToInjectBasedOnPolicy returns additional input parameters,
// which have to be injected into the Action, based on the current policies.
// TODO: check if rules has AdditionalInput to inject and if implementation expects AdditionalInput
func (e *PolicyEnforcedClient) ListAdditionalInputToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicgraphql.ImplementationRevision) map[string]interface{} {
	if policyRule.Inject == nil ||
		len(policyRule.Inject.AdditionalInput) == 0 ||
		implRev.Spec == nil ||
		implRev.Spec.AdditionalInput == nil ||
		implRev.Spec.AdditionalInput.Parameters == nil {
		return nil
	}
	return policyRule.Inject.AdditionalInput
}

// FindInterfaceRevision finds InterfaceRevision for the provided reference.
// It will return nil, if no revision was found.
func (e *PolicyEnforcedClient) FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference) (*hubpublicgraphql.InterfaceRevision, error) {
	return e.hubCli.FindInterfaceRevision(ctx, ref)
}

// SetPolicyOrder sets the policy merging order for the client. This setter is thread safe.
func (e *PolicyEnforcedClient) SetPolicyOrder(order policy.MergeOrder) {
	e.mu.Lock()
	e.policyOrder = order
	e.mergePolicies()
	e.mu.Unlock()
}

// SetGlobalPolicy sets global policy to use. This setter is thread safe.
func (e *PolicyEnforcedClient) SetGlobalPolicy(p policy.Policy) {
	e.mu.Lock()
	e.globalPolicy = p
	e.mergePolicies()
	e.mu.Unlock()
}

// SetActionPolicy sets policy to use during actiom workflow rendering. This setter is thread safe.
func (e *PolicyEnforcedClient) SetActionPolicy(p policy.ActionPolicy) {
	e.mu.Lock()
	e.actionPolicy = policy.Policy(p)
	e.mergePolicies()
	e.mu.Unlock()
}

// PushWorkflowStepPolicy adds a workflow policy to use during rendering a step. This setter is thread safe.
func (e *PolicyEnforcedClient) PushWorkflowStepPolicy(workflowPolicy policy.WorkflowPolicy) error {
	e.mu.Lock()
	p, err := workflowPolicy.ToPolicy()
	if err != nil {
		return errors.Wrap(err, "while getting Policy from WorkflowPolicy")
	}
	e.workflowStepPolicies = append(e.workflowStepPolicies, p)
	e.mergePolicies()
	e.mu.Unlock()
	return nil
}

// PopWorkflowStepPolicy removes latest added workflow policy. This setter is thread safe.
func (e *PolicyEnforcedClient) PopWorkflowStepPolicy() {
	e.mu.Lock()
	if len(e.workflowStepPolicies) > 0 {
		e.workflowStepPolicies = e.workflowStepPolicies[:len(e.workflowStepPolicies)-1]
	}
	e.mergePolicies()
	e.mu.Unlock()
}

// Policy gets policy which the Client uses. This getter is thread safe.
func (e *PolicyEnforcedClient) Policy() policy.Policy {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.mergedPolicy
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

	for i := range in {
		typeRef := in[i]
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
