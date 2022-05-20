package printer

import (
	"io"

	"k8s.io/cli-runtime/pkg/printers"
)

// JSONPath prints selected data from JSON.
type JSONPath struct {
	printer *printers.JSONPathPrinter
}

// Print executes k8s JSONPath printer to write to given writer selected data from JSON.
func (p *JSONPath) Print(in interface{}, w io.Writer) error {
	p.printer.EnableJSONOutput(true)

	return p.printer.JSONPath.Execute(w, in)
}
