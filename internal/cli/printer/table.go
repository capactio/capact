package printer

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// TableData holds data specific values that should be printed in table format.
type TableData struct {
	Headers      []string
	MultipleRows [][]string
	SingleRow    []string
}

var _ Printer = &Table{}

// TableDataProvider provides data specific values.
type TableDataProvider func(in interface{}) (TableData, error)

// Table prints data in table format.
type Table struct {
	dataProvider TableDataProvider
}

// Print creates table with provided data and writes it to a given writer.
func (p *Table) Print(in interface{}, w io.Writer) error {
	data, err := p.dataProvider(in)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(data.Headers)
	table.SetAutoWrapText(true)
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	table.SetRowLine(true)

	if len(data.MultipleRows) > 0 {
		table.AppendBulk(data.MultipleRows)
	}

	if len(data.SingleRow) > 0 {
		table.Append(data.SingleRow)
	}

	table.Render()

	return nil
}
