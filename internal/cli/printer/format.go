package printer

import (
	"fmt"
)

// Format is a type for capturing supported output formats.
// Implements pflag.Value interface.
type PrintFormat string

// ErrInvalidFormatType is returned when an unsupported format type is used
var ErrInvalidFormatType = fmt.Errorf("invalid output format type")

const (
	TableFormat PrintFormat = "table"
	JSONFormat  PrintFormat = "json"
	YAMLFormat  PrintFormat = "yaml"
)

// String returns the string representation of the Format
func (o PrintFormat) String() string {
	return string(o)
}

// String returns the string representation of the Format
func (o PrintFormat) IsValid() bool {
	switch o {
	case TableFormat, JSONFormat, YAMLFormat:
		return true
	}
	return false
}

func (o *PrintFormat) Set(in string) error {
	*o = PrintFormat(in)
	if !o.IsValid() {
		return ErrInvalidFormatType
	}
	return nil
}

func (o *PrintFormat) Type() string {
	return "string"
}
