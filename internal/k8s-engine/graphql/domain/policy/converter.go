package policy

import (
	"encoding/json"

	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) FromGraphQLInput(in graphql.PolicyInput) clusterpolicy.ClusterPolicy {
	var rules clusterpolicy.RulesList

	for _, gqlRule := range in.Rules {
		rules = append(rules, clusterpolicy.RulesForInterface{
			Interface: c.manifestRefFromGraphQLInput(gqlRule.Interface),
			OneOf:     c.policyRulesFromGraphQLInput(gqlRule.OneOf),
		})
	}

	return clusterpolicy.ClusterPolicy{
		APIVersion: clusterpolicy.CurrentAPIVersion,
		Rules:      rules,
	}
}

func (c *Converter) ToGraphQL(in clusterpolicy.ClusterPolicy) graphql.Policy {
	var gqlRules []*graphql.RulesForInterface

	for _, rule := range in.Rules {
		gqlRules = append(gqlRules, &graphql.RulesForInterface{
			Interface: c.manifestRefToGraphQL(rule.Interface),
			OneOf:     c.policyRulesToGraphQL(rule.OneOf),
		})
	}

	return graphql.Policy{
		Rules: gqlRules,
	}
}

func (c *Converter) policyRulesToGraphQL(in []clusterpolicy.Rule) []*graphql.PolicyRule {
	var gqlRules []*graphql.PolicyRule

	for _, rule := range in {
		gqlRule := &graphql.PolicyRule{
			ImplementationConstraints: &graphql.PolicyRuleImplementationConstraints{
				Requires:   c.manifestRefsToGraphQL(rule.ImplementationConstraints.Requires),
				Attributes: c.manifestRefsToGraphQL(rule.ImplementationConstraints.Attributes),
				Path:       rule.ImplementationConstraints.Path,
			},
			Inject: c.policyInjectDataToGraphQL(rule.Inject),
		}

		gqlRules = append(gqlRules, gqlRule)
	}

	return gqlRules
}

func (c *Converter) policyInjectDataToGraphQL(data *clusterpolicy.InjectData) *graphql.PolicyRuleInjectData {
	if data == nil {
		return nil
	}

	return &graphql.PolicyRuleInjectData{
		TypeInstances:   c.typeInstancesToInjectToGraphQL(data.TypeInstances),
		AdditionalInput: data.AdditionalInput,
	}
}

func (c *Converter) policyRulesFromGraphQLInput(in []*graphql.PolicyRuleInput) []clusterpolicy.Rule {
	var rules []clusterpolicy.Rule

	for _, gqlRule := range in {
		var implConstraints clusterpolicy.ImplementationConstraints
		if gqlRule.ImplementationConstraints != nil {
			implConstraints = clusterpolicy.ImplementationConstraints{
				Requires:   c.manifestRefsFromGraphQLInput(gqlRule.ImplementationConstraints.Requires),
				Attributes: c.manifestRefsFromGraphQLInput(gqlRule.ImplementationConstraints.Attributes),
				Path:       gqlRule.ImplementationConstraints.Path,
			}
		}

		rule := clusterpolicy.Rule{
			ImplementationConstraints: implConstraints,
			Inject:                    c.policyInjectDataFromGraphQLInput(gqlRule.Inject),
		}

		rules = append(rules, rule)
	}

	return rules
}

func (c *Converter) policyInjectDataFromGraphQLInput(input *graphql.PolicyRuleInjectDataInput) *clusterpolicy.InjectData {
	if input == nil {
		return nil
	}

	var additionalInput interface{}

	if input.AdditionalInput != nil {
		if err := json.Unmarshal([]byte(*input.AdditionalInput), &additionalInput); err != nil {
			// TODO: handle the error better
			additionalInput = nil
		}
	}

	return &clusterpolicy.InjectData{
		TypeInstances:   c.typeInstancesToInjectFromGraphQLInput(input.TypeInstances),
		AdditionalInput: additionalInput,
	}
}

func (c *Converter) manifestRefToGraphQL(in types.ManifestRef) *graphql.ManifestReferenceWithOptionalRevision {
	return &graphql.ManifestReferenceWithOptionalRevision{
		Path:     in.Path,
		Revision: in.Revision,
	}
}

func (c *Converter) manifestRefsToGraphQL(in *[]types.ManifestRef) []*graphql.ManifestReferenceWithOptionalRevision {
	if in == nil {
		return nil
	}

	var out []*graphql.ManifestReferenceWithOptionalRevision
	for _, item := range *in {
		out = append(out, c.manifestRefToGraphQL(item))
	}

	return out
}

func (c *Converter) typeInstancesToInjectToGraphQL(in []clusterpolicy.TypeInstanceToInject) []*graphql.TypeInstanceReference {
	var out []*graphql.TypeInstanceReference

	for _, item := range in {
		out = append(out, &graphql.TypeInstanceReference{
			ID:      item.ID,
			TypeRef: c.manifestRefToGraphQL(item.TypeRef),
		})
	}

	return out
}

func (c *Converter) manifestRefFromGraphQLInput(in *graphql.ManifestReferenceInput) types.ManifestRef {
	if in == nil {
		return types.ManifestRef{}
	}

	return types.ManifestRef{
		Path:     in.Path,
		Revision: in.Revision,
	}
}

func (c *Converter) manifestRefsFromGraphQLInput(in []*graphql.ManifestReferenceInput) *[]types.ManifestRef {
	if in == nil {
		return nil
	}

	var out []types.ManifestRef
	for _, item := range in {
		if item == nil {
			continue
		}

		out = append(out, c.manifestRefFromGraphQLInput(item))
	}

	return &out
}

func (c *Converter) typeInstancesToInjectFromGraphQLInput(in []*graphql.TypeInstanceReferenceInput) []clusterpolicy.TypeInstanceToInject {
	if in == nil {
		return nil
	}

	var out []clusterpolicy.TypeInstanceToInject
	for _, item := range in {
		out = append(out, clusterpolicy.TypeInstanceToInject{
			ID:      item.ID,
			TypeRef: c.manifestRefFromGraphQLInput(item.TypeRef),
		})
	}

	return out
}
