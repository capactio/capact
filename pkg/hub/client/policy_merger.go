package client

import (
	"reflect"

	"capact.io/capact/internal/maps"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
)

func (e *PolicyEnforcedClient) mergePolicies() {
	currentPolicy := policy.Policy{}

	for _, p := range e.policyOrder {
		switch p {
		case policy.Global:
			applyInterfacePolicy(&currentPolicy.Interface, e.globalPolicy.Interface)
			applyTypeInstancePolicy(&currentPolicy.TypeInstance, e.globalPolicy.TypeInstance)
		case policy.Action:
			applyInterfacePolicy(&currentPolicy.Interface, e.actionPolicy.Interface)
			applyTypeInstancePolicy(&currentPolicy.TypeInstance, e.actionPolicy.TypeInstance)
		case policy.Workflow:
			for _, wp := range e.workflowStepPolicies {
				// ignore TypeInstance Policy on Workflow as it's not supported,
				// see: policy.WorkflowPolicy type.
				applyInterfacePolicy(&currentPolicy.Interface, wp.Interface)
			}
		}
	}
	e.mergedPolicy = currentPolicy
}

// RequiredTypeInstancesForRule returns the merged list of TypeInstances from Rule and Defaults.
func (e *PolicyEnforcedClient) RequiredTypeInstancesForRule(policyRule policy.Rule) []policy.RequiredTypeInstanceToInject {
	mergedPolicy := e.Policy()
	// prefer policy Rule over Default
	return mergeRequiredTypeInstances(mergedPolicy.Interface.DefaultRequiredTypeInstancesToInject(), policyRule.RequiredTypeInstancesToInject())
}

func applyInterfacePolicy(currentPolicy *policy.InterfacePolicy, newPolicy policy.InterfacePolicy) {
	// Default
	if len(newPolicy.DefaultRequiredTypeInstancesToInject()) > 0 {
		if currentPolicy.Default == nil || currentPolicy.Default.Inject == nil {
			currentPolicy.Default = &policy.InterfaceDefault{
				Inject: &policy.DefaultInject{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{},
				},
			}
		}
		currentPolicy.Default.Inject.RequiredTypeInstances = mergeRequiredTypeInstances(currentPolicy.Default.Inject.RequiredTypeInstances, newPolicy.Default.Inject.RequiredTypeInstances)
	}

	// from new policy we are checking if there are the same rules. If yes we fill missing data,
	// if not we add a rule to the end
	// current policy is a higher priority policy
	for _, newRuleForInterface := range newPolicy.Rules {
		policyRuleIndex := getIndexOfInterfacePolicyRule(currentPolicy, newRuleForInterface)
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
	if newRule.Inject.AdditionalParameters != nil {
		rule.Inject.AdditionalParameters = mergeAdditionalParameters(newRule.Inject.AdditionalParameters, rule.Inject.AdditionalParameters)
	}
	// merge RequiredTypeInstances
	if newRule.Inject.RequiredTypeInstances != nil {
		rule.Inject.RequiredTypeInstances = mergeRequiredTypeInstances(newRule.Inject.RequiredTypeInstances, rule.Inject.RequiredTypeInstances)
	}
	// merge AdditionalTypeInstances
	if newRule.Inject.AdditionalTypeInstances != nil {
		rule.Inject.AdditionalTypeInstances = mergeAdditionalTypeInstances(newRule.Inject.AdditionalTypeInstances, rule.Inject.AdditionalTypeInstances)
	}
}

func getIndexOfInterfacePolicyRule(p *policy.InterfacePolicy, rule policy.RulesForInterface) int {
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

func mergeRequiredTypeInstances(current, overwrite []policy.RequiredTypeInstanceToInject) []policy.RequiredTypeInstanceToInject {
	out := append([]policy.RequiredTypeInstanceToInject{}, current...)
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

func mergeAdditionalTypeInstances(current, overwrite []policy.AdditionalTypeInstanceToInject) []policy.AdditionalTypeInstanceToInject {
	out := append([]policy.AdditionalTypeInstanceToInject{}, current...)
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

func mergeAdditionalParameters(current, overwrite []policy.AdditionalParametersToInject) []policy.AdditionalParametersToInject {
	out := append([]policy.AdditionalParametersToInject{}, current...)
	for _, newParam := range overwrite {
		found := false
		for i, param := range current {
			if newParam.Name == param.Name {
				found = true
				out[i] = policy.AdditionalParametersToInject{
					Name:  param.Name,
					Value: maps.Merge(param.Value, newParam.Value),
				}
			}
		}
		if !found {
			out = append(out, newParam)
		}
	}
	return out
}

func applyTypeInstancePolicy(currentPolicy *policy.TypeInstancePolicy, newPolicy policy.TypeInstancePolicy) {
	for _, newRule := range newPolicy.Rules {
		policyRuleIndex := getIndexOfTypeInstancePolicyRule(currentPolicy, newRule)
		if policyRuleIndex == -1 {
			newRuleForInterface := newRule.DeepCopy()
			currentPolicy.Rules = append(currentPolicy.Rules, *newRuleForInterface)
			continue
		}

		// override
		currentPolicy.Rules[policyRuleIndex] = newRule
	}
}

func getIndexOfTypeInstancePolicyRule(current *policy.TypeInstancePolicy, newRule policy.RulesForTypeInstance) int {
	for i, currentRule := range current.Rules {
		if currentRule.TypeRef.Path != newRule.TypeRef.Path {
			continue
		}
		curRev := ptr.StringPtrToString(currentRule.TypeRef.Revision)
		newRev := ptr.StringPtrToString(newRule.TypeRef.Revision)

		if curRev != newRev {
			continue
		}

		return i
	}

	return -1
}
