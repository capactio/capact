package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogrusSpinnerFormatter implements a naive support for spinner for logrus logger.
// If message starts with a gerund then it's spinner is active until next message is logged.
type LogrusSpinnerFormatter struct {
	failedPreviously bool
	spinner          *Status
}

// NewLogrusSpinnerFormatter returns a new LogrusSpinnerFormatter instance.
func NewLogrusSpinnerFormatter(header string) *LogrusSpinnerFormatter {
	return &LogrusSpinnerFormatter{
		spinner: NewStatus(os.Stdout, header),
	}
}

// Format formats logrus entry
func (f *LogrusSpinnerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.spinner.End(!f.failedPreviously)
	f.spinner.Step(fmt.Sprintf(entry.Message))

	switch entry.Level {
	case logrus.DebugLevel, logrus.InfoLevel, logrus.TraceLevel:
		f.failedPreviously = false
	case logrus.PanicLevel, logrus.FatalLevel:
		f.spinner.End(false)
	default:
		f.failedPreviously = true
	}

	// Here is a naive assumption that if the first message word is not a gerund,
	// it's not a long-running task and can be already marked as done.
	// In the worst case, we will mark as done a message indicating long-running task a bit to fast.
	words := strings.Fields(entry.Message)
	if isEmpty(words) || isNotGerund(words[0]) {
		f.spinner.End(!f.failedPreviously)
	}

	return []byte{}, nil
}

func isEmpty(in []string) bool {
	return len(in) == 0
}
func isNotGerund(in string) bool {
	return !strings.HasSuffix(in, "ing")
}
