package printer

import (
	"io"

	"k8s.io/cli-runtime/pkg/printers"
)

type JSONPath struct{}

func (p *JSONPath) Print(in interface{}, w io.Writer) error {
	printer, err := printers.NewJSONPathPrinter(".[0].latestResourceVersion.spec.value") // TODO Move init somewhere else
	if err != nil {
		return err
	}

	return printer.JSONPath.Execute(w, in)
}
