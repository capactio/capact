package printer

import (
	"fmt"
	"io"
	"time"

	"capact.io/capact/internal/cli"

	"github.com/fatih/color"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Spinner defines interface for terminal spinner.
type Spinner interface {
	Start(stage string)
	Active() bool
	Stop(msg string)
}

// Status provides functionality to display steps progress in terminal.
type Status struct {
	w io.Writer

	spinner         Spinner
	durationSprintf func(format string, a ...interface{}) string

	timeStarted time.Time
	stage       string
}

// NewStatus returns a new Status instance.
func NewStatus(w io.Writer, header string) *Status {
	if header != "" {
		fmt.Fprintln(w, header)
	}

	st := &Status{
		w: w,
	}
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

	started := ""
	if cli.VerboseMode.IsEnabled() {
		s.timeStarted = time.Now()
		started = fmt.Sprintf(" [started %s]", s.timeStarted.Format("15:04 MST"))
	}

	s.stage = fmt.Sprintf(stageFmt, args...)
	msg := fmt.Sprintf("%s%s", s.stage, started)
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

// Writer returns underlying io.Writer
func (s *Status) Writer() io.Writer {
	return s.w
}
