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
	writer io.Writer
	output PrintFormat

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
}

// PrintFormat returns default print format type.
func (r *ResourcePrinter) PrintFormat() PrintFormat {
	return r.output
}

// Print prints received object in requested format.
func (r *ResourcePrinter) Print(in interface{}) error {
	printer, err := r.getPrinter()
	if err != nil {
		return err
	}

	return printer.Print(in, r.writer)
}

func (r *ResourcePrinter) getPrinter() (Printer, error) {
	var printFormat PrintFormat

	if strings.HasPrefix(string(r.output), string(JSONPathFormat)) {
		printFormat = JSONPathFormat
	} else {
		printFormat = r.output
	}

	printer, found := r.printers[printFormat]
	if !found {
		return nil, fmt.Errorf("printer %q is not available", r.output)
	}

	if printFormat == JSONPathFormat {
		templatePrefix := string(JSONPathFormat) + "="

		if !strings.HasPrefix(string(r.output), string(templatePrefix)) {
			return nil, fmt.Errorf("JSON path output template should be prefixed with %q", templatePrefix)
		}

		template := string(r.output)[len(templatePrefix):]
		fmt.Println(template)

		jsonPathPrinter, err := printers.NewJSONPathPrinter(template)
		if err != nil {
			return nil, errors.Wrap(err, "while creating JSON path printer")
		}

		return &JSONPath{printer: jsonPathPrinter}, nil
	}

	return printer, nil
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
