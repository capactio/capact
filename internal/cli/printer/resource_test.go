package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourcePrinter_Print_Success(t *testing.T) {
	tests := []struct {
		name             string
		printerOption    ResourcePrinterOption
		printFormat      PrintFormat
		jsonPathTemplate string // used only for JSON path print format
		data             map[string]string
		expected         string
	}{
		{
			name:          "resource printer with json should print json when json output format is selected",
			printerOption: WithJSON(),
			printFormat:   JSONFormat,
			data:          map[string]string{"foo": "bar"},
			expected:      "{\n  \"foo\": \"bar\"\n}",
		},
		{
			name:             "resource printer with json path should print part of json when json path output format is selected and template is specified",
			printerOption:    WithJSONPath(),
			printFormat:      JSONPathFormat,
			jsonPathTemplate: "{.foo}",
			data:             map[string]string{"foo": "bar"},
			expected:         "[\n    \"bar\"\n]\n",
		},
		{
			name:          "resource printer with yaml should print yaml when yaml output format is selected",
			printerOption: WithYAML(),
			printFormat:   YAMLFormat,
			data:          map[string]string{"foo": "bar"},
			expected:      "foo: bar\n",
		},
		{
			name: "resource printer with table should print table when table output format is selected",
			printerOption: WithTable(func(_ interface{}) (TableData, error) {
				return TableData{
					Headers:   []string{"foo"},
					SingleRow: []string{"bar"},
				}, nil
			}),
			printFormat: TableFormat,
			expected:    "  FOO  \n-------\n  bar  \n-------\n",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var writer bytes.Buffer
			resourcePrinter := NewForResource(&writer, test.printerOption)
			resourcePrinter.output = test.printFormat
			resourcePrinter.template = test.jsonPathTemplate

			resourcePrinter.Print(test.data)

			assert.EqualValues(t, test.expected, writer.String())
		})
	}
}

func TestResourcePrinter_Print_Failure(t *testing.T) {
	tests := []struct {
		name             string
		printerOption    ResourcePrinterOption
		printFormat      PrintFormat
		jsonPathTemplate string // used only for JSON path
	}{
		{
			name:          "resource printer should return error when unknown print format is selected",
			printerOption: WithJSON(),
			printFormat:   "unknown print format",
		},
		{
			name:          "resource printer with json path should return error when json path print format is selected and template is not specified",
			printerOption: WithJSONPath(),
			printFormat:   JSONPathFormat,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var writer bytes.Buffer
			resourcePrinter := NewForResource(&writer, test.printerOption)
			resourcePrinter.output = test.printFormat
			resourcePrinter.template = test.jsonPathTemplate

			err := resourcePrinter.Print(struct{}{})

			assert.Error(t, err)
		})
	}
}
