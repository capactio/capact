package printer

import (
	"io"

	"sigs.k8s.io/yaml"
)

var _ Printer = &YAML{}

type YAML struct{}

func (p *YAML) Print(in interface{}, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
