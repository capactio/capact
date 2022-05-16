package printer

import (
	"io"

	"k8s.io/cli-runtime/pkg/printers"
)

type JSONPath struct {
	printer *printers.JSONPathPrinter
}

func (p *JSONPath) Print(in interface{}, w io.Writer) error {
	return p.printer.JSONPath.Execute(w, in)
}
