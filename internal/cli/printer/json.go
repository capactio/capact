package printer

import (
	"io"

	"github.com/hokaccha/go-prettyjson"
)

var _ Printer = &JSON{}

type JSON struct{}

func (p *JSON) Print(in interface{}, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}
