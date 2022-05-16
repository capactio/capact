package printer

import (
	"io"

	"k8s.io/cli-runtime/pkg/printers"
)

type JSONPath struct{}

func (p *JSONPath) Print(in interface{}, w io.Writer) error {
	printer, err := printers.NewJSONPathPrinter("{[0].backend.id}") // TODO Move somewhere else
	if err != nil {
		return err
	}

	return printer.JSONPath.Execute(w, in)
}
