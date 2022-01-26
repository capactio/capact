package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"k8s.io/utils/strings/slices"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
)

type validateFn func(ctx context.Context, entity types.Implementation) (ValidationResult, error)

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
	results := ValidationResult{}
	var entity types.Implementation
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Implementation type")
	}
	validateFns := []validateFn{v.validateInputArtifactsNames, v.checkManifestRevisionsExist}

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
	var manifestRefsToCheck []hubpublicgraphql.ManifestReference

	// Attributes
	for path, attr := range entity.Metadata.Attributes {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
			Path:     path,
			Revision: attr.Revision,
		})
	}

	// AdditionalParameters
	if entity.Spec.AdditionalInput != nil {
		// Parameters
		for _, param := range entity.Spec.AdditionalInput.Parameters {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(param.TypeRef))
		}

		// TypeInstances
		for _, ti := range entity.Spec.AdditionalInput.TypeInstances {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(ti.TypeRef))
		}
	}

	// AdditionalOutput
	if entity.Spec.AdditionalOutput != nil {
		for _, ti := range entity.Spec.AdditionalOutput.TypeInstances {
			if ti.TypeRef == nil {
				continue
			}

			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(*ti.TypeRef))
		}
	}

	// Implements
	for _, implementsItem := range entity.Spec.Implements {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(implementsItem))
	}

	// Requires
	for requiresKey, requiresValue := range entity.Spec.Requires {
		var itemsToCheck []types.RequireEntity
		itemsToCheck = append(itemsToCheck, requiresValue.OneOf...)
		itemsToCheck = append(itemsToCheck, requiresValue.AllOf...)
		itemsToCheck = append(itemsToCheck, requiresValue.AnyOf...)

		for _, requiresSubItem := range itemsToCheck {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
				Path:     strings.Join([]string{requiresKey, requiresSubItem.Name}, "."),
				Revision: requiresSubItem.Revision,
			})
		}
	}

	// Imports
	for _, importsItem := range entity.Spec.Imports {
		for _, method := range importsItem.Methods {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
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
		interfaceInput, err := v.fetchInterfaceInput(ctx, hubpublicgraphql.InterfaceReference{
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
	workflow, err := decodeImplArgsToArgoWorkflow(entity.Spec.Action.Args)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while decoding Implementation arguments to Argo workflow")
	}
	if workflow != nil && workflow.WorkflowSpec != nil {
		idx, err := argo.GetEntrypointWorkflowIndex(workflow)
		if err != nil {
			return ValidationResult{}, errors.Wrap(err, "while getting entrypoint index from workflow")
		}
		workflowArtifacts = append(workflowArtifacts, workflow.Templates[idx].Inputs.Artifacts...)
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

func (v *RemoteImplementationValidator) fetchInterfaceInput(ctx context.Context, interfaceRef hubpublicgraphql.InterfaceReference, hub Hub) (hubpublicgraphql.InterfaceInput, error) {
	iface, err := hub.FindInterfaceRevision(ctx, interfaceRef, public.WithInterfaceRevisionFields(public.InterfaceRevisionInputFields))
	if err != nil {
		return hubpublicgraphql.InterfaceInput{}, errors.Wrap(err, "while looking for Interface definition")
	}
	if iface == nil {
		return hubpublicgraphql.InterfaceInput{}, fmt.Errorf("interface %s:%s was not found in Hub", interfaceRef.Path, interfaceRef.Revision)
	}

	if iface.Spec == nil || iface.Spec.Input == nil {
		return hubpublicgraphql.InterfaceInput{}, nil
	}

	return *iface.Spec.Input, nil
}
func decodeImplArgsToArgoWorkflow(implArgs map[string]interface{}) (*argo.Workflow, error) {
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

// Name returns the validator name.
func (v *RemoteImplementationValidator) Name() string {
	return "RemoteImplementationValidator"
}
