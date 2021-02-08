package argo

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
