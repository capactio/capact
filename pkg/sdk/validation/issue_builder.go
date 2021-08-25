package validation

import (
	"fmt"
	"strings"
	"sync"

	"capact.io/capact/internal/multierror"
	gomultierror "github.com/hashicorp/go-multierror"
)

// IssueBuilder provides functionality to report issue by name and return aggregated result.
type IssueBuilder struct {
	mux    sync.Mutex
	issues Result
	header string
}

// NewResultBuilder returns a new IssueBuilder instance.
func NewResultBuilder(header string) *IssueBuilder {
	return &IssueBuilder{
		issues: Result{},
		header: header,
	}
}

// ReportIssue reports issue by a given field name.
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

// Result returns validation result index by field name.
func (bldr *IssueBuilder) Result() Result {
	if bldr == nil {
		return nil
	}

	for field, issues := range bldr.issues {
		if issues == nil {
			continue
		}
		issues.ErrorFormat = headeredErrListFormatFunc(fmt.Sprintf("- %s %q", bldr.header, field))
	}

	return bldr.issues
}

// headeredErrListFormatFunc is a basic formatter that outputs the errors as
// a bullet point list with a given header.
func headeredErrListFormatFunc(fieldName string) gomultierror.ErrorFormatFunc {
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
