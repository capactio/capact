package metadata

import (
	"fmt"
	"strings"

	"capact.io/capact/pkg/engine/k8s/policy"
)

type typeInstanceKind string

const (
	requiredTypeInstance   typeInstanceKind = "RequiredTypeInstance"
	additionalTypeInstance typeInstanceKind = "AdditionalTypeInstance"
	backendTypeInstance    typeInstanceKind = "BackendTypeInstance"
)

// TypeInstanceMetadata defines metadata for required and additional TypeInstances defined in Policy.
type TypeInstanceMetadata struct {
	ID          string
	Name        *string
	Description *string
	Kind        typeInstanceKind
}

// String returns string with details of a given TypeInstanceMetadata.
func (m TypeInstanceMetadata) String(withKind bool) string {
	tiDetails := []string{
		fmt.Sprintf("ID: %q", m.ID),
	}

	if m.Name != nil {
		tiDetails = append(tiDetails, fmt.Sprintf("name: %q", *m.Name))
	}

	if m.Description != nil {
		tiDetails = append(tiDetails, fmt.Sprintf("description: %q", *m.Description))
	}

	detailsStr := strings.Join(tiDetails, ", ")
	if withKind {
		return fmt.Sprintf("%s (%s)", m.Kind, detailsStr)
	}

	return detailsStr
}

// TypeInstanceIDsWithUnresolvedMetadataForPolicy filters TypeInstances that have unresolved metadata.
func TypeInstanceIDsWithUnresolvedMetadataForPolicy(in policy.Policy) []TypeInstanceMetadata {
	var tis []TypeInstanceMetadata

	// Interface
	for _, rule := range in.Interface.Rules {
		for _, ruleItem := range rule.OneOf {
			tis = append(tis, TypeInstanceIDsWithUnresolvedMetadataForRule(ruleItem)...)
		}
	}

	// TypeInstances backends
	for _, rule := range in.TypeInstance.Rules {
		if rule.Backend.TypeRef != nil && rule.Backend.TypeRef.Path != "" && rule.Backend.TypeRef.Revision != "" {
			continue
		}

		tis = append(tis, TypeInstanceMetadata{
			ID:          rule.Backend.ID,
			Description: rule.Backend.Description,
			Kind:        backendTypeInstance,
		})
	}

	return tis
}

// TypeInstanceIDsWithUnresolvedMetadataForRule filters TypeInstances that have unresolved metadata for a given rule.
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
