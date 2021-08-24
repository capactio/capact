package policy

import (
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// Converter provides functionality to convert GraphQL DTO to models.
type Converter struct{}

// NewConverter returns an new Converter instance.
func NewConverter() *Converter {
	return &Converter{}
}

// FromGraphQLInput coverts Graphql Policy data to model.
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
		Rules: rules,
	}, nil
}

// ToGraphQL converts Policy model representation to GraphQL DTO.
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
		RequiredTypeInstances: c.typeInstancesToInjectToGraphQL(data.RequiredTypeInstances),
		AdditionalParameters:  c.additionalParametersToInjectToGraphQL(data.AdditionalParameters),
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

	params, err := c.additionalParametersToInjectFromGraphQLInput(input.AdditionalParameters)
	if err != nil {
		return nil, errors.Wrap(err, "while converting additional parameters")
	}

	return &policy.InjectData{
		RequiredTypeInstances: c.typeInstancesToInjectFromGraphQLInput(input.RequiredTypeInstances),
		AdditionalParameters:  params,
	}, nil
}

func (c *Converter) manifestRefToGraphQL(in types.ManifestRefWithOptRevision) *graphql.ManifestReferenceWithOptionalRevision {
	return &graphql.ManifestReferenceWithOptionalRevision{
		Path:     in.Path,
		Revision: in.Revision,
	}
}

func (c *Converter) manifestRefsToGraphQL(in *[]types.ManifestRefWithOptRevision) []*graphql.ManifestReferenceWithOptionalRevision {
	if in == nil {
		return nil
	}

	var out []*graphql.ManifestReferenceWithOptionalRevision
	for _, item := range *in {
		out = append(out, c.manifestRefToGraphQL(item))
	}

	return out
}

func (c *Converter) typeInstancesToInjectToGraphQL(in []policy.RequiredTypeInstanceToInject) []*graphql.RequiredTypeInstanceReference {
	var out []*graphql.RequiredTypeInstanceReference

	for _, item := range in {
		out = append(out, &graphql.RequiredTypeInstanceReference{
			ID:          item.ID,
			Description: item.Description,
		})
	}

	return out
}

func (c *Converter) additionalParametersToInjectToGraphQL(in []policy.AdditionalParametersToInject) []*graphql.AdditionalParameter {
	var out []*graphql.AdditionalParameter

	for _, item := range in {
		out = append(out, &graphql.AdditionalParameter{
			Name:  item.Name,
			Value: item.Value,
		})
	}

	return out
}

func (c *Converter) manifestRefFromGraphQLInput(in *graphql.ManifestReferenceInput) types.ManifestRefWithOptRevision {
	if in == nil {
		return types.ManifestRefWithOptRevision{}
	}

	return types.ManifestRefWithOptRevision{
		Path:     in.Path,
		Revision: in.Revision,
	}
}

func (c *Converter) manifestRefsFromGraphQLInput(in []*graphql.ManifestReferenceInput) *[]types.ManifestRefWithOptRevision {
	if in == nil {
		return nil
	}

	var out []types.ManifestRefWithOptRevision
	for _, item := range in {
		if item == nil {
			continue
		}

		out = append(out, c.manifestRefFromGraphQLInput(item))
	}

	return &out
}

func (c *Converter) typeInstancesToInjectFromGraphQLInput(in []*graphql.RequiredTypeInstanceReferenceInput) []policy.RequiredTypeInstanceToInject {
	if in == nil {
		return nil
	}

	var out []policy.RequiredTypeInstanceToInject
	for _, item := range in {
		out = append(out, policy.RequiredTypeInstanceToInject{
			RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
				ID:          item.ID,
				Description: item.Description,
			},
		})
	}

	return out
}

func (c *Converter) additionalParametersToInjectFromGraphQLInput(in []*graphql.AdditionalParameterInput) ([]policy.AdditionalParametersToInject, error) {
	var out []policy.AdditionalParametersToInject
	for _, item := range in {
		additionalInput, ok := item.Value.(map[string]interface{})
		if !ok {
			return nil, ErrCannotConvertAdditionalInput
		}
		out = append(out, policy.AdditionalParametersToInject{
			Name:  item.Name,
			Value: additionalInput,
		})
	}

	return out, nil
}
