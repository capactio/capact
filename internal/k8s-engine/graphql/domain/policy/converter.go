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
	ifaceRules, err := c.interfaceFromGraphQLInput(in.Interface)
	if err != nil {
		return policy.Policy{}, err
	}

	typeInstanceRules := c.typeInstanceFromGraphQLInput(in.TypeInstance)

	return policy.Policy{
		Interface:    ifaceRules,
		TypeInstance: typeInstanceRules,
	}, nil
}

func (c *Converter) interfaceFromGraphQLInput(in *graphql.InterfacePolicyInput) (policy.InterfacePolicy, error) {
	if in == nil {
		return policy.InterfacePolicy{}, nil
	}
	var rules policy.InterfaceRulesList

	for _, gqlRule := range in.Rules {
		iface := c.manifestRefFromGraphQLInput(gqlRule.Interface)
		policyRules, err := c.policyRulesFromGraphQLInput(gqlRule.OneOf)
		if err != nil {
			return policy.InterfacePolicy{}, errors.Wrapf(err, "while converting 'OneOf' rules for %q", iface.String())
		}

		rules = append(rules, policy.RulesForInterface{
			Interface: iface,
			OneOf:     policyRules,
		})
	}

	var interfaceDefaults *policy.InterfaceDefault
	if in.Default != nil && in.Default.Inject != nil {
		interfaceDefaults = &policy.InterfaceDefault{
			Inject: &policy.DefaultInject{
				RequiredTypeInstances: c.requiredTypeInstancesToInjectFromGraphQLInput(in.Default.Inject.RequiredTypeInstances),
			},
		}
	}

	return policy.InterfacePolicy{
		Default: interfaceDefaults,
		Rules:   rules,
	}, nil
}

func (c *Converter) typeInstanceFromGraphQLInput(in *graphql.TypeInstancePolicyInput) policy.TypeInstancePolicy {
	if in == nil {
		return policy.TypeInstancePolicy{}
	}
	var rules []policy.RulesForTypeInstance

	for _, gqlRule := range in.Rules {
		gqlRef, gqlBackend := gqlRule.TypeRef, gqlRule.Backend
		if gqlRef == nil || gqlBackend == nil {
			continue
		}

		ref := types.ManifestRefWithOptRevision(*gqlRef)
		rules = append(rules, policy.RulesForTypeInstance{
			TypeRef: ref,
			Backend: policy.TypeInstanceBackend{
				TypeInstanceReference: policy.TypeInstanceReference{
					ID:          gqlBackend.ID,
					Description: gqlBackend.Description,
				},
			},
		})
	}

	return policy.TypeInstancePolicy{Rules: rules}
}

// ToGraphQL converts Policy model representation to GraphQL DTO.
func (c *Converter) ToGraphQL(in policy.Policy) graphql.Policy {
	return graphql.Policy{
		Interface:    c.interfaceToGraphQL(in.Interface),
		TypeInstance: c.typeInstanceToGraphQL(in.TypeInstance),
	}
}

func (c *Converter) typeInstanceToGraphQL(in policy.TypeInstancePolicy) *graphql.TypeInstancePolicy {
	var gqlRules []*graphql.RulesForTypeInstance

	for _, rule := range in.Rules {
		ref := graphql.ManifestReferenceWithOptionalRevision(rule.TypeRef)

		gqlRules = append(gqlRules, &graphql.RulesForTypeInstance{
			TypeRef: &ref,
			Backend: &graphql.TypeInstanceBackendRule{
				ID:          rule.Backend.ID,
				Description: rule.Backend.Description,
			},
		})
	}

	return &graphql.TypeInstancePolicy{
		Rules: gqlRules,
	}
}

func (c *Converter) interfaceToGraphQL(in policy.InterfacePolicy) *graphql.InterfacePolicy {
	var gqlRules []*graphql.RulesForInterface

	for _, rule := range in.Rules {
		gqlRules = append(gqlRules, &graphql.RulesForInterface{
			Interface: c.manifestRefToGraphQL(rule.Interface),
			OneOf:     c.policyRulesToGraphQL(rule.OneOf),
		})
	}

	var defaultForInterface *graphql.DefaultForInterface
	if in.DefaultRequiredTypeInstancesToInject() != nil {
		defaultForInterface = &graphql.DefaultForInterface{
			Inject: &graphql.DefaultInjectForInterface{
				RequiredTypeInstances: c.requiredTypeInstancesToInjectToGraphQL(in.Default.Inject.RequiredTypeInstances),
			},
		}
	}

	return &graphql.InterfacePolicy{
		Default: defaultForInterface,
		Rules:   gqlRules,
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
		RequiredTypeInstances:   c.requiredTypeInstancesToInjectToGraphQL(data.RequiredTypeInstances),
		AdditionalParameters:    c.additionalParametersToInjectToGraphQL(data.AdditionalParameters),
		AdditionalTypeInstances: c.additionalTypeInstancesToInjectToGraphQL(data.AdditionalTypeInstances),
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
		RequiredTypeInstances:   c.requiredTypeInstancesToInjectFromGraphQLInput(input.RequiredTypeInstances),
		AdditionalParameters:    params,
		AdditionalTypeInstances: c.additionalTypeInstancesToInjectFromGraphQLInput(input.AdditionalTypeInstances),
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

func (c *Converter) requiredTypeInstancesToInjectToGraphQL(in []policy.RequiredTypeInstanceToInject) []*graphql.RequiredTypeInstanceReference {
	var out []*graphql.RequiredTypeInstanceReference

	for _, item := range in {
		out = append(out, &graphql.RequiredTypeInstanceReference{
			ID:          item.ID,
			Description: item.Description,
		})
	}

	return out
}

func (c *Converter) additionalTypeInstancesToInjectToGraphQL(in []policy.AdditionalTypeInstanceToInject) []*graphql.AdditionalTypeInstanceReference {
	var out []*graphql.AdditionalTypeInstanceReference

	for _, item := range in {
		out = append(out, &graphql.AdditionalTypeInstanceReference{
			ID:   item.ID,
			Name: item.Name,
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

func (c *Converter) requiredTypeInstancesToInjectFromGraphQLInput(in []*graphql.RequiredTypeInstanceReferenceInput) []policy.RequiredTypeInstanceToInject {
	if in == nil {
		return nil
	}

	var out []policy.RequiredTypeInstanceToInject
	for _, item := range in {
		out = append(out, policy.RequiredTypeInstanceToInject{
			TypeInstanceReference: policy.TypeInstanceReference{
				ID:          item.ID,
				Description: item.Description,
			},
		})
	}

	return out
}

func (c *Converter) additionalTypeInstancesToInjectFromGraphQLInput(in []*graphql.AdditionalTypeInstanceReferenceInput) []policy.AdditionalTypeInstanceToInject {
	if in == nil {
		return nil
	}

	var out []policy.AdditionalTypeInstanceToInject
	for _, item := range in {
		out = append(out, policy.AdditionalTypeInstanceToInject{
			AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
				Name: item.Name,
				ID:   item.ID,
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
