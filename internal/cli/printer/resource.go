package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/printers"
)

// Printer is an interface that knows how to print objects.
type Printer interface {
	// Print receives an object, formats it and prints it to a writer.
	Print(in interface{}, w io.Writer) error
}

// ResourcePrinter provides functionality to print a given resource in requested format.
// Can be configured with pflag.FlagSet.
type ResourcePrinter struct {
	writer   io.Writer
	output   PrintFormat
	template string

	printers map[PrintFormat]Printer
}

// NewForResource returns a new ResourcePrinter instance.
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

// ResourcePrinterOption allows ResourcePrinter instance customization.
type ResourcePrinterOption func(*ResourcePrinter)

// WithJSON registers JSON format type.
func WithJSON() ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[JSONFormat] = &JSON{}
	}
}

// WithJSONPath registers JSON path format type.
func WithJSONPath() ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		// Cannot fully initialize JSONPath because template at this time is unknown
		r.printers[JSONPathFormat] = &JSONPath{}
	}
}

// WithYAML registers YAML format type.
func WithYAML() ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[YAMLFormat] = &YAML{}
	}
}

// WithTable registers table format type.
func WithTable(provider TableDataProvider) ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.printers[TableFormat] = &Table{dataProvider: provider}
	}
}

// WithDefaultOutputFormat sets a default format type.
func WithDefaultOutputFormat(format PrintFormat) ResourcePrinterOption {
	return func(r *ResourcePrinter) {
		r.output = format
	}
}

// RegisterFlags registers ResourcePrinter terminal flags.
func (r *ResourcePrinter) RegisterFlags(flags *pflag.FlagSet) {
	flags.VarP(&r.output, "output", "o", fmt.Sprintf("Output format. One of: %s", r.availablePrinters()))

	if _, ok := r.printers[JSONPathFormat]; ok {
		flags.StringVarP(&r.template, "template", "t", "", "JSON path output template (https://kubernetes.io/docs/reference/kubectl/jsonpath)")
	}
}

// PrintFormat returns default print format type.
func (r *ResourcePrinter) PrintFormat() PrintFormat {
	return r.output
}

// Print prints received object in requested format.
func (r *ResourcePrinter) Print(in interface{}) error {
	printer, found := r.printers[r.output]
	if !found {
		return fmt.Errorf("printer %q is not available", r.output)
	}

	if r.output == JSONPathFormat {
		if r.template == "" {
			return errors.New("JSON path template is not specified")
		}

		jsonPathPrinter, err := printers.NewJSONPathPrinter(r.template)
		if err != nil {
			return errors.Wrap(err, "while creating JSON path printer")
		}

		printer = &JSONPath{printer: jsonPathPrinter}
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
