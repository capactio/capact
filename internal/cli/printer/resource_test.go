package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourcePrinterWithJsonPrintsJsonWhenJsonOutputFormatIsSelected(t *testing.T) {
	var writer bytes.Buffer
	resourcePrinter := NewForResource(&writer, WithJSON())
	resourcePrinter.output = JSONFormat

	resourcePrinter.Print(map[string]string{"foo": "bar"})

	assert.JSONEq(t, `{"foo": "bar"}`, writer.String())
}

func TestResourcePrinterWithJsonPathPrintsJsonPartWhenJsonPathOutputFormatIsSelected(t *testing.T) {
	var writer bytes.Buffer
	resourcePrinter := NewForResource(&writer, WithJSONPath())
	resourcePrinter.output = JSONPathFormat
	resourcePrinter.template = "{.foo}"

	resourcePrinter.Print(map[string]string{"foo": "bar"})

	assert.EqualValues(t, "[\n    \"bar\"\n]\n", writer.String())
}

func TestResourcePrinterWithYamlPrintsYamlWhenYamlOutputFormatIsSelected(t *testing.T) {
	var writer bytes.Buffer
	resourcePrinter := NewForResource(&writer, WithYAML())
	resourcePrinter.output = YAMLFormat

	resourcePrinter.Print(map[string]string{"foo": "bar"})

	assert.EqualValues(t, "foo: bar\n", writer.String())
}

func TestResourcePrinterWithTablePrintsTableWhenTableOutputFormatIsSelected(t *testing.T) {
	tableDataProvider := func(_ interface{}) (TableData, error) {
		return TableData{
			Headers:   []string{"foo"},
			SingleRow: []string{"bar"},
		}, nil
	}

	var writer bytes.Buffer
	resourcePrinter := NewForResource(&writer, WithTable(tableDataProvider))
	resourcePrinter.output = TableFormat

	resourcePrinter.Print(map[string]string{}) // data is created in tableDataProvider

	assert.EqualValues(t, "  FOO  \n-------\n  bar  \n-------\n", writer.String())
}

func TestResourcePrinterReturnsErrorWhenOutputFormatPrinterDoesNotExists(t *testing.T) {
	var writer bytes.Buffer
	resourcePrinter := NewForResource(&writer, WithJSON())
	resourcePrinter.output = "not existing output format"

	err := resourcePrinter.Print(map[string]string{})

	assert.Error(t, err)
}
