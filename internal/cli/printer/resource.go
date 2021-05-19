package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

// Printer is an interface that knows how to print objects.
type Printer interface {
	// Print receives an object, formats it and prints it to a writer.
	Print(in interface{}, w io.Writer) error
}

type ResourcePrinter struct {
	writer io.Writer
	output PrintFormat

	printers map[PrintFormat]Printer
}

func NewForResource(w io.Writer, opts ...ResourcePrinterOption) *ResourcePrinter {
	p := &ResourcePrinter{
		writer:   w,
		printers: map[PrintFormat]Printer{},
		output:   TableFormat,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type ResourcePrinterOption func(*ResourcePrinter)

func WithJSON() ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[JSONFormat] = &JSON{}
	}
}

func WithYAML() ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[YAMLFormat] = &YAML{}
	}
}

func WithTable(provider TableDataProvider) ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[TableFormat] = &Table{dataProvider: provider}
	}
}

func WithDefaultOutputFormat(format PrintFormat) ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.output = format
	}
}

func (r *ResourcePrinter) RegisterFlags(flags *pflag.FlagSet) {
	flags.VarP(&r.output, "output", "o", fmt.Sprintf("Output format. One of: %s", r.availablePrinters()))
}

func (r *ResourcePrinter) Print(in interface{}) error {
	printer, found := r.printers[r.output]
	if !found {
		return fmt.Errorf("printer %q is not available", r.output)
	}

	return printer.Print(in, r.writer)
}

func (r *ResourcePrinter) availablePrinters() string {
	var out []string
	for key := range r.printers {
		out = append(out, key.String())
	}

	// We generate doc automatically, so it needs to be deterministic
	sort.Strings(out)

	return strings.Join(out, " | ")
}
