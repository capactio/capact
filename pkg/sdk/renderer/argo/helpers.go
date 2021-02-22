package argo

import (
	"context"
	"fmt"
	"strings"

	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

func interfaceRefToOCH(in types.InterfaceRef) ochpublicgraphql.InterfaceReference {
	return ochpublicgraphql.InterfaceReference{
		Path:     in.Path,
		Revision: stringOrEmpty(in.Revision),
	}
}

func stringOrEmpty(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func shouldExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
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

func findTypeInstanceTypeRef(typeInstanceName string, impl *ochpublicgraphql.ImplementationRevision, iface *ochpublicgraphql.InterfaceRevision) (*ochpublicgraphql.TypeReference, error) {
	toSearch := []*ochpublicgraphql.OutputTypeInstance{}

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
	if step.VoltronTypeInstanceOutputs != nil {
		for _, output := range step.VoltronTypeInstanceOutputs {
			if output.From == typeInstanceName {
				return &output
			}
		}
	}

	return nil
}

func argoArtifactRefToName(ref string) string {
	ref = strings.TrimPrefix(ref, "{{")
	ref = strings.TrimSuffix(ref, "}}")
	parts := strings.Split(ref, ".")

	prefix := parts[0]
	switch prefix {
	case "steps":
		stepName := parts[1]
		artifactName := parts[4]

		return fmt.Sprintf("%s-%s", stepName, artifactName)
	}

	return ""
}
