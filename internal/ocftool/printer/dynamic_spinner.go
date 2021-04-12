package printer

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
)

// DynamicSpinner is suitable for smart terminals which support animations (control cursor location, color etc.).
type DynamicSpinner struct {
	underlying *spinner.Spinner
}

func NewDynamicSpinner(w io.Writer) *DynamicSpinner {
	return &DynamicSpinner{
		underlying: spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(w)),
	}
}

func (d *DynamicSpinner) Start(stage string) {
	d.underlying.Prefix = " "
	d.underlying.Suffix = fmt.Sprintf(" %s", stage)
	d.underlying.Start()
}

func (d *DynamicSpinner) Active() bool {
	return d.underlying.Active()
}

func (d *DynamicSpinner) Stop(msg string) {
	d.underlying.FinalMSG = msg
	d.underlying.Stop()
}
