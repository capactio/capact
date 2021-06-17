package client

import "capact.io/capact/pkg/engine/k8s/policy"

func (e *PolicyEnforcedClient) mergePolicies() {
	newPolicy := policy.Policy{}

	for _, p := range e.policyOrder {
		if p == policy.Global {
			newPolicy = applyPolicy(newPolicy, e.globalPolicy)
		} else if p == policy.Action {
			newPolicy = applyPolicy(newPolicy, e.actionPolicy)
		} else if p == policy.Workflow {
			for _, wp := range e.workflowStepPolicy {
				newPolicy = applyPolicy(newPolicy, wp)
			}
		}
	}
	e.mergedPolicy = newPolicy
}

// from new policy we are checking if there are the same rules. If yes we fill missing data,
// if not we add a rule to the end
// current policy is a higher priority policy
func applyPolicy(currentPolicy, newPolicy policy.Policy) policy.Policy {
	for _, newRuleForInterface := range newPolicy.Rules {
		policyRuleIndex := getIndexOfPolicyRule(currentPolicy, newRuleForInterface)
		if policyRuleIndex == -1 {
			currentPolicy.Rules = append(currentPolicy.Rules, newRuleForInterface)
			continue
		}
		ruleForInterface := currentPolicy.Rules[policyRuleIndex]
		for _, newRule := range newRuleForInterface.OneOf {
			ruleIndex := getIndexOfOneOfRule(ruleForInterface.OneOf, newRule)
			if ruleIndex == -1 {
				currentPolicy.Rules[policyRuleIndex].OneOf = append(currentPolicy.Rules[policyRuleIndex].OneOf, newRule)
				continue
			}
			if newRule.Inject == nil {
				break
			}
			rule := ruleForInterface.OneOf[ruleIndex]
			// merge Additional Input
			if newRule.Inject.AdditionalInput != nil {
				ruleForInterface.OneOf[ruleIndex].Inject.AdditionalInput = mergeMaps(newRule.Inject.AdditionalInput, rule.Inject.AdditionalInput)
			}
			// merge TypeInstances
			if newRule.Inject.TypeInstances != nil {
				ruleForInterface.OneOf[ruleIndex].Inject.TypeInstances = mergeTypeInstances(newRule.Inject.TypeInstances, rule.Inject.TypeInstances)
			}
		}
	}
	return currentPolicy
}

func getIndexOfPolicyRule(p policy.Policy, rule policy.RulesForInterface) int {
	for i, ruleForInterface := range p.Rules {
		if isForSameInterface(ruleForInterface, rule) {
			return i
		}
	}
	return -1
}

func getIndexOfOneOfRule(rules []policy.Rule, rule policy.Rule) int {
	for i, r := range rules {
		if isSameOneOf(r, rule) {
			return i
		}
	}
	return -1
}

func isForSameInterface(p1, p2 policy.RulesForInterface) bool {
	if p1.Interface.Path != p2.Interface.Path {
		return false
	}
	return p1.Interface.Revision == p2.Interface.Revision
}

func isSameOneOf(a, b policy.Rule) bool {
	return a.ImplementationConstraints.Path == b.ImplementationConstraints.Path
	//TODO match rules with different constraints
}

func mergeMaps(current, overwrite map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(current))
	for k, v := range current {
		out[k] = v
	}
	for k, v := range overwrite {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
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
