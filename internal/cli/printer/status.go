package printer

import (
	"fmt"
	"io"

	"capact.io/capact/internal/cli"

	"github.com/fatih/color"
)

// Spinner defines interface for terminal spinner.
type Spinner interface {
	Start(stage string)
	Active() bool
	Stop(msg string)
}

// Status provides functionality to display steps progress in terminal.
type Status struct {
	stage   string
	spinner Spinner
}

// NewStatus returns a new Status instance.
func NewStatus(w io.Writer, header string) *Status {
	if header != "" {
		fmt.Fprintln(w, header)
	}

	st := &Status{}
	if cli.IsSmartTerminal(w) {
		st.spinner = NewDynamicSpinner(w)
	} else {
		st.spinner = NewStaticSpinner(w)
	}

	return st
}

// Step starts spinner for a given step.
func (s *Status) Step(stageFmt string, args ...interface{}) {
	// Finish previously started step
	s.End(true)

	s.stage = fmt.Sprintf(stageFmt, args...)
	s.spinner.Start(s.stage)
}

// End marks started step as completed.
func (s *Status) End(success bool) {
	if !s.spinner.Active() {
		return
	}

	var finalMsg string
	if success {
		finalMsg = fmt.Sprintf(" %s %s\n", color.GreenString("✓"), s.stage)
	} else {
		finalMsg = fmt.Sprintf(" %s %s\n", color.RedString("✗"), s.stage)
	}

	s.spinner.Stop(finalMsg)
}
