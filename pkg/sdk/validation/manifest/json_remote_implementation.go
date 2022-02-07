package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	typesutil "capact.io/capact/pkg/hub/client/public/facade/types"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	"capact.io/capact/pkg/sdk/renderer/argo"
	"k8s.io/utils/strings/slices"

	"github.com/dustin/go-humanize/english"
	"github.com/pkg/errors"
)

// ParentNodesAssociation represents relations between parent node and associated other types.
// - key holds the parent node path
// - value holds list of associated Types
type ParentNodesAssociation map[string][]types.TypeRef

// RemoteImplementationValidator is a validator for Implementation manifest, which calls Hub in order to do validation checks.
type RemoteImplementationValidator struct {
	hub Hub
}

type validateFn func(ctx context.Context, entity types.Implementation) (ValidationResult, error)

// NewRemoteImplementationValidator creates new RemoteImplementationValidator.
func NewRemoteImplementationValidator(hub Hub) *RemoteImplementationValidator {
	return &RemoteImplementationValidator{
		hub: hub,
	}
}

// Do is a method which triggers the validation.
func (v *RemoteImplementationValidator) Do(ctx context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	results := ValidationResult{}
	var entity types.Implementation
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Implementation type")
	}
	validateFns := []validateFn{
		v.checkManifestRevisionsExist,
		v.checkRequiresParentNodes,
		v.validateInputArtifactsNames,
	}

	for _, fn := range validateFns {
		validationResults, err := fn(ctx, entity)
		if err != nil {
			return ValidationResult{}, err
		}
		results.Errors = append(results.Errors, validationResults.Errors...)
	}

	return results, nil
}

func (v *RemoteImplementationValidator) checkManifestRevisionsExist(ctx context.Context, entity types.Implementation) (ValidationResult, error) {
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
	for requiresKey, reqItem := range entity.Spec.Requires {
		typesThatShouldExist, _ := v.resolveRequiresPath(requiresKey, reqItem)
		manifestRefsToCheck = append(manifestRefsToCheck, typesThatShouldExist...)
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
	return checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
}

func (v *RemoteImplementationValidator) validateInputArtifactsNames(ctx context.Context, entity types.Implementation) (ValidationResult, error) {
	var validationErrs []error
	var interfacesInputNames []string
	var implAdditionalInput []string
	var workflowArtifacts []wfv1.Artifact

	//1. get interface input names
	for _, implementsItem := range entity.Spec.Implements {
		interfaceInput, err := v.fetchInterfaceInput(ctx, gqlpublicapi.InterfaceReference{
			Path:     implementsItem.Path,
			Revision: implementsItem.Revision,
		}, v.hub)
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while fetching Interface inputs")
		}
		for _, inputParameter := range interfaceInput.Parameters {
			interfacesInputNames = append(interfacesInputNames, inputParameter.Name)
		}
		for _, inputTypeInstance := range interfaceInput.TypeInstances {
			interfacesInputNames = append(interfacesInputNames, inputTypeInstance.Name)
		}
	}

	//2. get implementation additional inputs
	if entity.Spec.AdditionalInput != nil {
		for name := range entity.Spec.AdditionalInput.Parameters {
			implAdditionalInput = append(implAdditionalInput, name)
		}
		for name := range entity.Spec.AdditionalInput.TypeInstances {
			implAdditionalInput = append(implAdditionalInput, name)
		}
	}

	//3. get inputs from entrypoint workflow template
	workflow, err := v.decodeImplArgsToArgoWorkflow(entity.Spec.Action.Args)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while decoding Implementation arguments to Argo workflow")
	}
	if workflow != nil && workflow.WorkflowSpec != nil {
		idx, err := argo.GetEntrypointWorkflowIndex(workflow)
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while getting entrypoint index from workflow")
		}
		workflowArtifacts = workflow.Templates[idx].Inputs.Artifacts
	}

	//4. verify if the inputs from Implementation and Interface match with Argo workflow artifacts
	for _, artifact := range workflowArtifacts {
		existsInInterface := slices.Contains(interfacesInputNames, artifact.Name)
		existsInAdditionalInput := slices.Contains(implAdditionalInput, artifact.Name)

		if existsInInterface &&
			artifact.Optional {
			validationErrs = append(validationErrs, fmt.Errorf("invalid workflow input artifact %q: it shouldn't be optional as it is defined as Interface input", artifact.Name))
		}
		if existsInAdditionalInput && !artifact.Optional {
			validationErrs = append(validationErrs, fmt.Errorf("invalid workflow input artifact %q: it should be optional, as it is defined as Implementation additional input", artifact.Name))
		}
		if !existsInInterface && !existsInAdditionalInput {
			validationErrs = append(validationErrs, fmt.Errorf("unknown workflow input artifact %q: there is no such input neither in Interface input, nor Implementation additional input", artifact.Name))
		}
	}

	return ValidationResult{Errors: validationErrs}, nil
}

func (v *RemoteImplementationValidator) fetchInterfaceInput(ctx context.Context, interfaceRef gqlpublicapi.InterfaceReference, hub Hub) (gqlpublicapi.InterfaceInput, error) {
	iface, err := hub.FindInterfaceRevision(ctx, interfaceRef, public.WithInterfaceRevisionFields(public.InterfaceRevisionInputFields))
	if err != nil {
		return gqlpublicapi.InterfaceInput{}, errors.Wrap(err, "while looking for Interface definition")
	}
	if iface == nil {
		return gqlpublicapi.InterfaceInput{}, fmt.Errorf("interface %s:%s was not found in Hub", interfaceRef.Path, interfaceRef.Revision)
	}

	if iface.Spec == nil || iface.Spec.Input == nil {
		return gqlpublicapi.InterfaceInput{}, nil
	}

	return *iface.Spec.Input, nil
}

func (v *RemoteImplementationValidator) decodeImplArgsToArgoWorkflow(implArgs map[string]interface{}) (*argo.Workflow, error) {
	var decodedImplArgs = struct {
		Workflow argo.Workflow `json:"workflow"`
	}{}

	b, err := json.Marshal(implArgs)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling Implementation arguments")
	}

	if err := json.Unmarshal(b, &decodedImplArgs); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling Implementation arguments to Argo Workflow")
	}
	return &decodedImplArgs.Workflow, nil
}

func (v *RemoteImplementationValidator) checkRequiresParentNodes(ctx context.Context, entity types.Implementation) (ValidationResult, error) {
	parentNodeTypesToCheck := ParentNodesAssociation{}
	for requiresKey, reqItem := range entity.Spec.Requires {
		_, typesThatHasParentNode := v.resolveRequiresPath(requiresKey, reqItem)

		for k, v := range typesThatHasParentNode {
			parentNodeTypesToCheck[k] = append(parentNodeTypesToCheck[k], v...)
		}
	}

	return v.checkParentNodesAssociation(ctx, parentNodeTypesToCheck)
}

// Name returns the validator name.
func (v *RemoteImplementationValidator) Name() string {
	return "RemoteImplementationValidator"
}

func (v *RemoteImplementationValidator) resolveRequiresPath(parentPrefix string, reqItem types.Require) ([]gqlpublicapi.ManifestReference, ParentNodesAssociation) {
	var (
		typesThatShouldExist   []gqlpublicapi.ManifestReference
		typesThatHasParentNode = ParentNodesAssociation{}
	)

	var allReqItems []types.RequireEntity
	allReqItems = append(allReqItems, reqItem.OneOf...)
	allReqItems = append(allReqItems, reqItem.AllOf...)
	allReqItems = append(allReqItems, reqItem.AnyOf...)

	for _, requiresSubItem := range allReqItems {
		ref := types.TypeRef{
			Path:     strings.Join([]string{parentPrefix, requiresSubItem.Name}, "."), // default assumption
			Revision: requiresSubItem.Revision,
		}

		// Check if item under requires section is a concrete Type. If yes, it needs to be attached to the parent node.
		// For example:
		//   requires:
		//     cap.core.type.platform:
		//      oneOf:
		//        - name: cap.type.platform.nomad # this MUST be attached to `cap.core.type.platform`
		//          revision: 0.1.0
		if strings.HasPrefix(requiresSubItem.Name, types.OCFPathPrefix) {
			ref.Path = requiresSubItem.Name
			typesThatHasParentNode[parentPrefix] = append(typesThatHasParentNode[parentPrefix], ref)
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
	for parentNode, expTypesRefs := range relations {
		gotAttachedTypes, err := typesutil.ListAdditionalRefs(ctx, v.hub, expTypesRefs)
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while fetching Types based on parent node")
		}

		missingEntries := v.detectMissingChildren(gotAttachedTypes, expTypesRefs, parentNode)
		if len(missingEntries) == 0 {
			continue
		}
		validationErrs = append(validationErrs, fmt.Errorf("%s %s %s not attached to %q parent node",
			english.PluralWord(len(missingEntries), "Type", ""),
			english.WordSeries(missingEntries, "and"),
			english.PluralWord(len(missingEntries), "is", "are"),
			parentNode,
		))
	}

	return ValidationResult{Errors: validationErrs}, nil
}

func (v *RemoteImplementationValidator) detectMissingChildren(gotAttachedTypes typesutil.ListAdditionalRefsOutput, expAttachedTypes []types.TypeRef, expParent string) []string {
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

		missingChildren = append(missingChildren, fmt.Sprintf(`"%s:%s"`, exp.Path, exp.Revision))
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
