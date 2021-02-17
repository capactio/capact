package argo

import (
	"context"

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

func findTypeInstanceTypeRef(typeInstanceName string, impl *ochpublicgraphql.ImplementationRevision, iface *ochpublicgraphql.InterfaceRevision) *ochpublicgraphql.TypeReference {
	if iface.Spec.Output != nil {
		for i := range iface.Spec.Output.TypeInstances {
			ti := iface.Spec.Output.TypeInstances[i]
			if ti.Name == typeInstanceName {
				return ti.TypeRef
			}
		}
	}

	if impl.Spec.AdditionalOutput != nil {
		for i := range impl.Spec.AdditionalOutput.TypeInstances {
			ti := impl.Spec.AdditionalOutput.TypeInstances[i]

			if ti.Name == typeInstanceName {
				return ti.TypeRef
			}
		}
	}

	return nil
}
