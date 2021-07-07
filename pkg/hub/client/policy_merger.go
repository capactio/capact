package client

import (
	"reflect"

	tools "capact.io/capact/internal"

	"capact.io/capact/pkg/engine/k8s/policy"
)

func (e *PolicyEnforcedClient) mergePolicies() {
	currentPolicy := policy.Policy{}

	for _, p := range e.policyOrder {
		switch p {
		case policy.Global:
			applyPolicy(&currentPolicy, e.globalPolicy)
		case policy.Action:
			applyPolicy(&currentPolicy, e.actionPolicy)
		case policy.Workflow:
			for _, wp := range e.workflowStepPolicies {
				applyPolicy(&currentPolicy, wp)
			}
		}
	}
	e.mergedPolicy = currentPolicy
}

// from new policy we are checking if there are the same rules. If yes we fill missing data,
// if not we add a rule to the end
// current policy is a higher priority policy
func applyPolicy(currentPolicy *policy.Policy, newPolicy policy.Policy) {
	for _, newRuleForInterface := range newPolicy.Rules {
		policyRuleIndex := getIndexOfPolicyRule(currentPolicy, newRuleForInterface)
		if policyRuleIndex == -1 {
			newRuleForInterface := newRuleForInterface.DeepCopy()
			currentPolicy.Rules = append(currentPolicy.Rules, *newRuleForInterface)
			continue
		}
		ruleForInterface := currentPolicy.Rules[policyRuleIndex]
		for _, newRule := range newRuleForInterface.OneOf {
			ruleIndex := getIndexOfOneOfRule(ruleForInterface.OneOf, newRule)
			if ruleIndex == -1 {
				newRule := newRule.DeepCopy()
				currentPolicy.Rules[policyRuleIndex].OneOf = append(currentPolicy.Rules[policyRuleIndex].OneOf, *newRule)
				continue
			}
			mergeRules(&ruleForInterface.OneOf[ruleIndex], newRule)
		}
	}
}

func mergeRules(rule *policy.Rule, newRule policy.Rule) {
	if newRule.Inject == nil {
		return
	}
	if rule.Inject == nil {
		rule.Inject = &policy.InjectData{}
	}
	// merge Additional Input
	if newRule.Inject.AdditionalInput != nil {
		newMap := tools.MergeMaps(newRule.Inject.AdditionalInput, rule.Inject.AdditionalInput)
		rule.Inject.AdditionalInput = newMap
	}
	// merge TypeInstances
	if newRule.Inject.TypeInstances != nil {
		rule.Inject.TypeInstances = mergeTypeInstances(newRule.Inject.TypeInstances, rule.Inject.TypeInstances)
	}
}

func getIndexOfPolicyRule(p *policy.Policy, rule policy.RulesForInterface) int {
	for i, ruleForInterface := range p.Rules {
		if isForSameInterface(ruleForInterface, rule) {
			return i
		}
	}
	return -1
}

func getIndexOfOneOfRule(rules []policy.Rule, rule policy.Rule) int {
	for i, r := range rules {
		if areImplementationConstraintsEqual(r, rule) {
			return i
		}
	}
	return -1
}

func isForSameInterface(p1, p2 policy.RulesForInterface) bool {
	if p1.Interface.Path != p2.Interface.Path {
		return false
	}

	var revision1, revision2 string
	if p1.Interface.Revision != nil {
		revision1 = *p1.Interface.Revision
	}
	if p2.Interface.Revision != nil {
		revision2 = *p2.Interface.Revision
	}

	return revision1 == revision2
}

func areImplementationConstraintsEqual(a, b policy.Rule) bool {
	return reflect.DeepEqual(a.ImplementationConstraints, b.ImplementationConstraints)
}

func mergeTypeInstances(current, overwrite []policy.TypeInstanceToInject) []policy.TypeInstanceToInject {
	out := append([]policy.TypeInstanceToInject{}, current...)
	for _, newTI := range overwrite {
		found := false
		for i, ti := range current {
			if newTI.TypeRef.Path == ti.TypeRef.Path && newTI.TypeRef.Revision == ti.TypeRef.Revision {
				found = true
				out[i] = newTI
			}
		}
		if !found {
			out = append(out, newTI)
		}
	}
	return out
}
