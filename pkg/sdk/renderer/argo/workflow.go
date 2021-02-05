package argo

func getEntrypointWorkflowIndex(w *Workflow) (int, bool) {
	if w == nil {
		return 0, false
	}
	for idx, tmpl := range w.Templates {
		if tmpl.Name == w.Entrypoint {
			return idx, true
		}
	}

	return 0, false
}
