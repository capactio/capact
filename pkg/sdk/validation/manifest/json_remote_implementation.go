package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

const (
	ocfPathPrefix       = "cap."
	typeListQueryFields = public.TypeRevisionRootFields | public.TypeRevisionSpecAdditionalRefsField
)

// ParentNodesAssociation represents relations between parent node and associated other types.
// - key holds the parent node path
// - value holds list of associated Types
type ParentNodesAssociation map[string][]types.ManifestRef

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
		ref := types.ManifestRef{
			Path:     strings.Join([]string{abstractPrefix, requiresSubItem.Name}, "."), // default assumption
			Revision: requiresSubItem.Revision,
		}

		// Check if item under requires section is a concrete Type. If yes, it needs to be attached to the parent node.
		// For example:
		//   requires:
		//     cap.core.type.platform:
		//      oneOf:
		//        - name: cap.type.platform.cloud-foundry # this MUST be attached to `cap.core.type.platform`
		//          revision: 0.1.0
		if strings.HasPrefix(requiresSubItem.Name, ocfPathPrefix) {
			ref.Path = requiresSubItem.Name
			typesThatHasParentNode[abstractPrefix] = append(typesThatHasParentNode[abstractPrefix], ref)
		}

		typesThatShouldExist = append(typesThatShouldExist, gqlpublicapi.ManifestReference(ref))
	}

	return typesThatShouldExist, typesThatHasParentNode
}

// checkParentNodesAssociation check whether a given Types is associated with a given parent node.
// BEWARE: Types not found in Hub are ignored.
func (v *RemoteImplementationValidator) checkParentNodesAssociation(ctx context.Context, relations ParentNodesAssociation) (ValidationResult, error) {
	if len(relations) == 0 {
		return ValidationResult{}, nil
	}

	var validationErrs []error
	for abstractNode, expTypesRefs := range relations {
		typesPath, expAttachedTypes := v.mapToPathAndPathRevIndex(expTypesRefs)

		filter := regexutil.OrStringSlice(typesPath)
		res, err := v.hub.ListTypes(ctx, public.WithTypeRevisions(typeListQueryFields), public.WithTypeFilter(gqlpublicapi.TypeFilter{
			PathPattern: ptr.String(filter),
		}))
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while fetching Types based on abstract node")
		}

		gotAttachedTypes := map[string][]string{}
		for _, item := range res {
			if item == nil {
				continue
			}
			for _, rev := range item.Revisions {
				if rev.Spec == nil {
					continue
				}
				gotAttachedTypes[v.key(item.Path, rev.Revision)] = rev.Spec.AdditionalRefs
			}
		}

		missingEntries := v.detectMissingChildren(gotAttachedTypes, expAttachedTypes, abstractNode)
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

func (v *RemoteImplementationValidator) mapToPathAndPathRevIndex(in []types.ManifestRef) ([]string, []string) {
	var (
		paths       []string
		pathsRevIdx []string
	)

	for _, expType := range in {
		paths = append(paths, expType.Path)
		pathsRevIdx = append(pathsRevIdx, v.key(expType.Path, expType.Revision))
	}

	return paths, pathsRevIdx
}

func (v *RemoteImplementationValidator) detectMissingChildren(gotAttachedTypes map[string][]string, expAttachedTypes []string, expParent string) []string {
	var missingChildren []string

	for _, exp := range expAttachedTypes {
		gotParents, found := gotAttachedTypes[exp]
		if !found {
			// Type not found in Hub, but it's not our job to report that
			continue
		}

		if v.stringSliceContains(gotParents, expParent) {
			continue
		}

		missingChildren = append(missingChildren, exp)
	}

	return missingChildren
}

func (v *RemoteImplementationValidator) stringSliceContains(slice []string, elem string) bool {
	for _, parent := range slice {
		if parent == elem {
			return true
		}
	}
	return false
}

func (v *RemoteImplementationValidator) key(a, b string) string {
	return fmt.Sprintf("%s:%s", a, b)
}
