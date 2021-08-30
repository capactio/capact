package metadata

import "capact.io/capact/pkg/engine/k8s/policy"



type typeInstanceKind string

const (
	requiredTypeInstance   typeInstanceKind = "RequiredTypeInstance"
	additionalTypeInstance typeInstanceKind = "AdditionalTypeInstance"
)

type TypeInstanceMetadata struct {
	ID          string
	Name        *string
	Description *string
	Kind        typeInstanceKind
}

func TypeInstanceIDsWithUnresolvedMetadataForPolicy(in policy.Policy) []TypeInstanceMetadata {
	var tis []TypeInstanceMetadata
	for _, rule := range in.Rules {
		for _, ruleItem := range rule.OneOf {
			tis = append(tis, TypeInstanceIDsWithUnresolvedMetadataForRule(ruleItem)...)
		}
	}

	return tis
}

func TypeInstanceIDsWithUnresolvedMetadataForRule(in policy.Rule) []TypeInstanceMetadata {
	if in.Inject == nil {
		return nil
	}

	var tis []TypeInstanceMetadata

	// Required TypeInstances
	for _, ti := range in.Inject.RequiredTypeInstances {
		if ti.TypeRef != nil && ti.TypeRef.Path != "" && ti.TypeRef.Revision != "" {
			continue
		}

		tis = append(tis, TypeInstanceMetadata{
			ID:          ti.ID,
			Description: ti.Description,
			Kind:        requiredTypeInstance,
		})
	}

	// Additional TypeInstances
	for _, ti := range in.Inject.AdditionalTypeInstances {
		if ti.TypeRef != nil && ti.TypeRef.Path != "" && ti.TypeRef.Revision != "" {
			continue
		}

		tiName := ti.Name
		tis = append(tis, TypeInstanceMetadata{
			ID:   ti.ID,
			Name: &tiName,
			Kind: additionalTypeInstance,
		})
	}

	return tis
}
