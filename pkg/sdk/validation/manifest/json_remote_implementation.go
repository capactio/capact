package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

const ocfPathPrefix = "cap."

// ParentNodesAssociation represents relations between parent node and associated other types.
// - key holds the parent node path
// - value holds list of associated Types
type ParentNodesAssociation map[string][]string

// RemoteImplementationValidator is a validator for Implementation manifest, which calls Hub in order to do validation checks.
type RemoteImplementationValidator struct {
	hub Hub
}

// NewRemoteImplementationValidator creates new RemoteImplementationValidator.
func NewRemoteImplementationValidator(hub Hub) *RemoteImplementationValidator {
	return &RemoteImplementationValidator{
		hub: hub,
	}
}

// Do is a method which triggers the validation.
func (v *RemoteImplementationValidator) Do(ctx context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Implementation
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Implementation type")
	}

	var manifestRefsToCheck []gqlpublicapi.ManifestReference

	// Attributes
	for path, attr := range entity.Metadata.Attributes {
		manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference{
			Path:     path,
			Revision: attr.Revision,
		})
	}

	// AdditionalParameters
	if entity.Spec.AdditionalInput != nil {
		// Parameters
		for _, param := range entity.Spec.AdditionalInput.Parameters {
			manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference(param.TypeRef))
		}

		// TypeInstances
		for _, ti := range entity.Spec.AdditionalInput.TypeInstances {
			manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference(ti.TypeRef))
		}
	}

	// AdditionalOutput
	if entity.Spec.AdditionalOutput != nil {
		for _, ti := range entity.Spec.AdditionalOutput.TypeInstances {
			if ti.TypeRef == nil {
				continue
			}

			manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference(*ti.TypeRef))
		}
	}

	// Implements
	for _, implementsItem := range entity.Spec.Implements {
		manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference(implementsItem))
	}

	// Requires
	parentNodeTypesToCheck := ParentNodesAssociation{}
	for requiresKey, reqItem := range entity.Spec.Requires {
		typesThatShouldExist, typesThatHasParentNode := v.resolveRequiresPath(requiresKey, reqItem)
		manifestRefsToCheck = append(manifestRefsToCheck, typesThatShouldExist...)

		for k, v := range typesThatHasParentNode {
			parentNodeTypesToCheck[k] = append(parentNodeTypesToCheck[k], v...)
		}
	}

	// Imports
	for _, importsItem := range entity.Spec.Imports {
		for _, method := range importsItem.Methods {
			manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference{
				Path:     strings.Join([]string{importsItem.InterfaceGroupPath, method.Name}, "."),
				Revision: method.Revision,
			})
		}
	}

	// TODO: refactor after https://github.com/capactio/capact/pull/610
	res, err := v.checkParentNodesAssociation(ctx, parentNodeTypesToCheck)
	if !res.Valid() || err != nil {
		return res, err
	}

	return checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
}

// Name returns the validator name.
func (v *RemoteImplementationValidator) Name() string {
	return "RemoteImplementationValidator"
}

func (v *RemoteImplementationValidator) resolveRequiresPath(abstractPrefix string, reqItem types.Require) ([]gqlpublicapi.ManifestReference, ParentNodesAssociation) {
	var (
		typesThatShouldExist   []gqlpublicapi.ManifestReference
		typesThatHasParentNode = ParentNodesAssociation{}
	)

	var allReqItems []types.RequireEntity
	allReqItems = append(allReqItems, reqItem.OneOf...)
	allReqItems = append(allReqItems, reqItem.AllOf...)
	allReqItems = append(allReqItems, reqItem.AnyOf...)

	for _, requiresSubItem := range allReqItems {
		path := strings.Join([]string{abstractPrefix, requiresSubItem.Name}, ".")

		// Check if item is concrete Type. If yes, it needs to be attached to parent node. For example:
		// requires:
		//   cap.core.type.platform:
		//    oneOf:
		//      - name: cap.type.platform.cloud-foundry # this MUST be attached to `cap.core.type.platform`
		//        revision: 0.1.0
		if strings.HasPrefix(requiresSubItem.Name, ocfPathPrefix) {
			path = requiresSubItem.Name
			typesThatHasParentNode[abstractPrefix] = append(typesThatHasParentNode[abstractPrefix], path)
		}

		typesThatShouldExist = append(typesThatShouldExist, gqlpublicapi.ManifestReference{
			Path:     path,
			Revision: requiresSubItem.Revision,
		})
	}

	return typesThatShouldExist, typesThatHasParentNode
}

// TODO: revisions
func (v *RemoteImplementationValidator) checkParentNodesAssociation(ctx context.Context, relations ParentNodesAssociation) (ValidationResult, error) {
	if len(relations) == 0 {
		return ValidationResult{}, nil
	}

	var validationErrs []error
	for abstractNode, expAttachedTypes := range relations {
		res, err := v.hub.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
			PathPattern: ptr.String(abstractNode),
		}))
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while fetching Types based on abstract node")
		}

		var gotAttachedTypes []string
		for _, item := range res {
			gotAttachedTypes = append(gotAttachedTypes, item.Path)
		}

		missingEntries := v.detectMissingEntriesInASet(gotAttachedTypes, expAttachedTypes)
		if len(missingEntries) == 0 {
			continue
		}

		validationErrs = append(validationErrs, fmt.Errorf("%s %s %s not attached to %s abstract node",
			english.PluralWord(len(missingEntries), "Type", ""),
			english.WordSeries(missingEntries, "and"),
			english.PluralWord(len(missingEntries), "is", "are"),
			abstractNode,
		))
	}

	return ValidationResult{Errors: validationErrs}, nil
}

func (v *RemoteImplementationValidator) detectMissingEntriesInASet(a, b []string) []string {
	// we don't do `len(a) != len(b)` as it only informs us that there will be definitely
	// some missing entries, but we need to find out names

	aSetIndex := make(map[string]struct{}, len(a))
	for _, val := range a {
		aSetIndex[val] = struct{}{}
	}

	var missingEntries []string
	for _, val := range b {
		if _, found := aSetIndex[val]; found {
			continue
		}

		missingEntries = append(missingEntries, val)
	}

	return missingEntries
}
