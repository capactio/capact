package validate

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type (
	Issue struct {
		Severity SeverityType // enum // default error
		Message  string
	}

	IssueBuilder struct {
		mux    sync.Mutex
		issues ValidationResult
		header string
	}
)

func NewResultBuilder(header string) *IssueBuilder {
	return &IssueBuilder{
		issues: ValidationResult{},
		header: header,
	}
}

type ValidationResult map[string]*multierror.Error

func (bldr *IssueBuilder) ReportIssue(field, format string, args ...interface{}) *IssueBuilder {
	if bldr == nil { // TODO: error?
		return nil
	}

	bldr.mux.Lock()
	defer bldr.mux.Unlock()
	err := fmt.Errorf(format, args...)
	bldr.issues[field] = multierror.Append(bldr.issues[field], err)

	return bldr
}

func (bldr *IssueBuilder) Result() ValidationResult {
	if bldr == nil {
		return nil
	}

	for field, issues := range bldr.issues {
		if issues == nil {
			continue
		}
		issues.ErrorFormat = ListFormatFunc(fmt.Sprintf("- %s %q", bldr.header, field))
	}

	return bldr.issues
}

func (issues ValidationResult) Len() int {
	cnt := 0
	for _, issues := range issues {
		if issues == nil {
			continue
		}
		cnt += issues.Len()
	}

	return cnt
}

func (issues ValidationResult) ErrorOrNil() error {
	var msgs []string
	for _, issues := range issues {
		if issues == nil {
			continue
		}
		msgs = append(msgs, issues.Error())
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

// ListFormatFunc is a basic formatter that outputs the number of errors
// that occurred along with a bullet point list of the errors.
func ListFormatFunc(fieldName string) multierror.ErrorFormatFunc {
	return func(es []error) string {
		points := make([]string, len(es))
		for i, err := range es {
			points[i] = fmt.Sprintf("* %s", err)
		}

		return fmt.Sprintf(
			"%s:\n    %s",
			fieldName, strings.Join(points, "\n    "))
	}
}

// maybe later

type ReportIssueOpt func(*Issue)

func WithSeverity(s SeverityType) ReportIssueOpt {
	return func(i *Issue) {
		i.Severity = s
	}
}

type SeverityType int

const (
	Error SeverityType = iota + 1
	Warning
)

func (s SeverityType) String() string {
	switch s {
	case Error:
		return "Error"
	case Warning:
		return "Warning"
	default:
		return ""
	}
}
