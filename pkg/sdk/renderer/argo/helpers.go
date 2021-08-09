package argo

import (
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

func interfaceRefToHub(in types.InterfaceRef) hubpublicgraphql.InterfaceReference {
	return hubpublicgraphql.InterfaceReference{
		Path:     in.Path,
		Revision: ptr.StringPtrToString(in.Revision),
	}
}

func getEntrypointWorkflowIndex(w *Workflow) (int, error) {
	if w == nil {
		return 0, NewWorkflowNilError()
	}

	for idx, tmpl := range w.Templates {
		if tmpl.Name == w.Entrypoint {
			return idx, nil
		}
	}

	return 0, NewEntrypointWorkflowIndexNotFoundError(w.Entrypoint)
}

func findTypeInstanceTypeRef(typeInstanceName string, impl *hubpublicgraphql.ImplementationRevision, iface *hubpublicgraphql.InterfaceRevision) (*hubpublicgraphql.TypeReference, error) {
	if iface == nil {
		return nil, NewTypeReferenceNotFoundError(typeInstanceName)
	}

	var toSearch []*hubpublicgraphql.OutputTypeInstance

	if iface.Spec.Output != nil {
		toSearch = append(toSearch, iface.Spec.Output.TypeInstances...)
	}

	if impl.Spec.AdditionalOutput != nil {
		toSearch = append(toSearch, impl.Spec.AdditionalOutput.TypeInstances...)
	}

	for i := range toSearch {
		ti := toSearch[i]
		if ti.Name == typeInstanceName {
			return ti.TypeRef, nil
		}
	}

	return nil, NewTypeReferenceNotFoundError(typeInstanceName)
}

func findOutputTypeInstance(step *WorkflowStep, typeInstanceName string) *TypeInstanceDefinition {
	for _, output := range step.CapactTypeInstanceOutputs {
		if output.From == typeInstanceName {
			return &output
		}
	}

	return nil
}

type argoArtifactRef struct {
	step string
	name string
}

// ArgoArtifactNoStep indicates that the Argo artifact was not produced in a workflow step.
const ArgoArtifactNoStep = ""

func getArgoArtifactRef(ref string) (*argoArtifactRef, error) {
	ref = strings.TrimPrefix(ref, "{{")
	ref = strings.TrimSuffix(ref, "}}")
	segments := strings.Split(ref, ".")

	invalidPathErrForRef := func(ref string, expectedSegments, actualSegments int) error {
		return fmt.Errorf("invalid artifact path '%s': expected %d path segments, instead got %d", ref, expectedSegments, actualSegments)
	}

	prefix := segments[0]
	switch prefix {
	case "steps":
		expectedSegments := 5
		if len(segments) < expectedSegments {
			return nil, invalidPathErrForRef(ref, expectedSegments, len(segments))
		}
		stepName := segments[1]
		artifactName := segments[4]
		return &argoArtifactRef{
			step: stepName,
			name: artifactName,
		}, nil
	case "inputs":
		expectedSegments := 3
		if len(segments) < expectedSegments {
			return nil, invalidPathErrForRef(ref, expectedSegments, len(segments))
		}
		artifactName := segments[2]
		return &argoArtifactRef{
			step: ArgoArtifactNoStep,
			name: artifactName,
		}, nil
	case "workflow":
		expectedSegments := 4
		if len(segments) < expectedSegments {
			return nil, invalidPathErrForRef(ref, expectedSegments, len(segments))
		}
		artifactName := segments[3]
		return &argoArtifactRef{
			step: ArgoArtifactNoStep,
			name: artifactName,
		}, nil
	}

	return nil, errors.New("not found")
}

func getAvailableTypeInstancesFromInputArtifacts(inputArtifacts []InputArtifact) map[argoArtifactRef]*string {
	availableTypeInstances := map[argoArtifactRef]*string{}

	for _, artifact := range inputArtifacts {
		if artifact.typeInstanceReference != nil {
			availableTypeInstances[argoArtifactRef{
				name: artifact.artifact.Name,
				step: ArgoArtifactNoStep,
			}] = artifact.typeInstanceReference
		}
	}

	return availableTypeInstances
}

func findInputArtifact(inputArtifacts []InputArtifact, name string) *InputArtifact {
	for _, art := range inputArtifacts {
		if art.artifact.Name == name {
			return &art
		}
	}

	return nil
}

func findTypeInstanceInputRef(refs []types.InputTypeInstanceRef, name string) *types.InputTypeInstanceRef {
	for i := range refs {
		ref := refs[i]
		if ref.Name == name {
			return &ref
		}
	}

	return nil
}

func addPrefix(prefix, s string) string {
	return fmt.Sprintf("%s-%s", prefix, s)
}
