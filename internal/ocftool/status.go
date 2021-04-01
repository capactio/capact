package ocftool

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
)

type StatusPrinter struct {
	stage         string
	spinner       *spinner.Spinner
	successFormat string
	failureFormat string
}

func NewStatusPrinter(w io.Writer, header string) *StatusPrinter {
	fmt.Fprintln(w, header)

	return &StatusPrinter{
		spinner:       spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(w)),
		successFormat: " \x1b[32m✓\x1b[0m %s\n",
		failureFormat: " \x1b[31m✗\x1b[0m %s\n",
	}
}

func (s *StatusPrinter) Step(stageFmt string, args ...interface{}) {
	stage := fmt.Sprintf(stageFmt, args...)

	// Finish previously started step
	if s.stage != "" {
		s.End(true)
	}

	s.stage = stage
	s.spinner.Prefix = " "
	s.spinner.Suffix = fmt.Sprintf(" %s", s.stage)
	s.spinner.Start()

}

func (s *StatusPrinter) End(success bool) {
	if !s.spinner.Active() {
		return
	}

	if success {
		s.spinner.FinalMSG = fmt.Sprintf(s.successFormat, s.stage)
	} else {
		s.spinner.FinalMSG = fmt.Sprintf(s.failureFormat, s.stage)
	}

	s.spinner.Stop()
}
