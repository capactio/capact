package policy

import (
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) FromGraphQLInput(in graphql.PolicyInput) (policy.Policy, error) {
	var rules policy.RulesList

	for _, gqlRule := range in.Rules {
		policyRules, err := c.policyRulesFromGraphQLInput(gqlRule.OneOf)
		if err != nil {
			return policy.Policy{}, errors.Wrap(err, "while getting Policy rules")
		}

		rules = append(rules, policy.RulesForInterface{
			Interface: c.manifestRefFromGraphQLInput(gqlRule.Interface),
			OneOf:     policyRules,
		})
	}

	return policy.Policy{
		APIVersion: policy.CurrentAPIVersion,
		Rules:      rules,
	}, nil
}

func (c *Converter) ToGraphQL(in policy.Policy) graphql.Policy {
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

func (c *Converter) policyRulesToGraphQL(in []policy.Rule) []*graphql.PolicyRule {
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

func (c *Converter) policyInjectDataToGraphQL(data *policy.InjectData) *graphql.PolicyRuleInjectData {
	if data == nil {
		return nil
	}

	return &graphql.PolicyRuleInjectData{
		TypeInstances:   c.typeInstancesToInjectToGraphQL(data.TypeInstances),
		AdditionalInput: data.AdditionalInput,
	}
}

func (c *Converter) policyRulesFromGraphQLInput(in []*graphql.PolicyRuleInput) ([]policy.Rule, error) {
	var rules []policy.Rule

	for _, gqlRule := range in {
		var implConstraints policy.ImplementationConstraints
		if gqlRule.ImplementationConstraints != nil {
			implConstraints = policy.ImplementationConstraints{
				Requires:   c.manifestRefsFromGraphQLInput(gqlRule.ImplementationConstraints.Requires),
				Attributes: c.manifestRefsFromGraphQLInput(gqlRule.ImplementationConstraints.Attributes),
				Path:       gqlRule.ImplementationConstraints.Path,
			}
		}

		injectData, err := c.policyInjectDataFromGraphQLInput(gqlRule.Inject)
		if err != nil {
			return nil, errors.Wrap(err, "while getting Policy inject data")
		}

		rule := policy.Rule{
			ImplementationConstraints: implConstraints,
			Inject:                    injectData,
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (c *Converter) policyInjectDataFromGraphQLInput(input *graphql.PolicyRuleInjectDataInput) (*policy.InjectData, error) {
	if input == nil {
		return nil, nil
	}

	var additionalInput map[string]interface{}

	if input.AdditionalInput != nil {
		var ok bool
		additionalInput, ok = input.AdditionalInput.(map[string]interface{})
		if !ok {
			return nil, ErrCannotConvertAdditionalInput
		}
	}

	return &policy.InjectData{
		TypeInstances:   c.typeInstancesToInjectFromGraphQLInput(input.TypeInstances),
		AdditionalInput: additionalInput,
	}, nil
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

func (c *Converter) typeInstancesToInjectToGraphQL(in []policy.TypeInstanceToInject) []*graphql.TypeInstanceReference {
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

func (c *Converter) typeInstancesToInjectFromGraphQLInput(in []*graphql.TypeInstanceReferenceInput) []policy.TypeInstanceToInject {
	if in == nil {
		return nil
	}

	var out []policy.TypeInstanceToInject
	for _, item := range in {
		out = append(out, policy.TypeInstanceToInject{
			ID:      item.ID,
			TypeRef: c.manifestRefFromGraphQLInput(item.TypeRef),
		})
	}

	return out
}
