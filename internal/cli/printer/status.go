package printer

import (
	"fmt"
	"io"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"

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
	stage           string
	spinner         Spinner
	timeStarted     time.Time
	durationSprintf func(format string, a ...interface{}) string
}

// NewStatus returns a new Status instance.
func NewStatus(w io.Writer, header string) *Status {
	if header != "" {
		fmt.Fprintln(w, header)
	}

	st := &Status{}
	if cli.IsSmartTerminal(w) {
		st.durationSprintf = color.New(color.Faint, color.Italic).Sprintf
		st.spinner = NewDynamicSpinner(w)
	} else {
		st.durationSprintf = fmt.Sprintf
		st.spinner = NewStaticSpinner(w)
	}

	return st
}

// Step starts spinner for a given step.
func (s *Status) Step(stageFmt string, args ...interface{}) {
	// Finish previously started step
	s.End(true)

	if cli.VerboseMode.IsEnabled() {
		s.timeStarted = time.Now()
	}

	s.stage = fmt.Sprintf(stageFmt, args...)
	msg := fmt.Sprintf("%s [started %s]", s.stage, s.timeStarted.Format("15:04 MST"))
	s.spinner.Start(msg)
}

// End marks started step as completed.
func (s *Status) End(success bool) {
	if !s.spinner.Active() {
		return
	}

	var icon string
	if success {
		icon = color.GreenString("✓")
	} else {
		icon = color.RedString("✗")
	}

	dur := ""
	if cli.VerboseMode.IsEnabled() {
		dur = s.durationSprintf(" [took %s]", duration.HumanDuration(time.Since(s.timeStarted)))
	}
	msg := fmt.Sprintf(" %s %s%s\n",
		icon, s.stage, dur)
	s.spinner.Stop(msg)
}
