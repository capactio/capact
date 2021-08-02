package printer

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type SpinnerFormatter struct {
	failedPreviously bool
	secondRun        bool
	spinner          *Status
}

func NewLogrusSpinnerFormatter(header string) *SpinnerFormatter {
	return &SpinnerFormatter{
		spinner: NewStatus(os.Stdout, header),
	}
}

func (f *SpinnerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.spinner.End(!f.failedPreviously)
	f.spinner.Step(entry.Message)

	if strings.EqualFold(entry.Message, "You can now use it like this:") {
		f.spinner.End(true)
	}
	switch entry.Level {
	case logrus.DebugLevel, logrus.InfoLevel, logrus.TraceLevel:
		f.failedPreviously = false
	case logrus.PanicLevel, logrus.FatalLevel:
		f.spinner.End(false)
	default:
		f.failedPreviously = true
	}

	return []byte{}, nil
}
