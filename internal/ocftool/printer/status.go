package printer

import (
	"fmt"
	"io"

	"capact.io/capact/internal/ocftool"

	"github.com/fatih/color"
)

type Spinner interface {
	Start(stage string)
	Active() bool
	Stop(msg string)
}

type Status struct {
	stage   string
	spinner Spinner
}

func NewStatus(w io.Writer, header string) *Status {
	fmt.Fprintln(w, header)

	st := &Status{}
	if ocftool.IsSmartTerminal(w) {
		st.spinner = NewDynamicSpinner(w)
	} else {
		st.spinner = NewStaticSpinner(w)
	}

	return st
}

func (s *Status) Step(stageFmt string, args ...interface{}) {
	// Finish previously started step
	s.End(true)

	s.stage = fmt.Sprintf(stageFmt, args...)
	s.spinner.Start(s.stage)
}

func (s *Status) End(success bool) {
	if !s.spinner.Active() {
		return
	}

	var finalMsg string
	if success {
		finalMsg = fmt.Sprintf(" %s %s", color.GreenString("✓"), s.stage)
	} else {
		finalMsg = fmt.Sprintf(" %s %s", color.RedString("✗"), s.stage)
	}

	s.spinner.Stop(finalMsg)
}
