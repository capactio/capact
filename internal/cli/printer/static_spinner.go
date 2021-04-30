package printer

import (
	"fmt"
	"io"
)

// StaticSpinner is suitable for non-smart terminals.
type StaticSpinner struct {
	w      io.Writer
	active bool
}

func (s *StaticSpinner) Start(stage string) {
	s.active = true
	fmt.Fprintf(s.w, " • %s\n", stage)
}

func (s *StaticSpinner) Active() bool {
	return s.active
}

func (s *StaticSpinner) Stop(msg string) {
	s.active = false
	fmt.Fprintln(s.w, msg)
}

func NewStaticSpinner(w io.Writer) *StaticSpinner {
	return &StaticSpinner{w: w}
}
