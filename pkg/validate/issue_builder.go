package validate

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type IssueBuilder struct {
	mux    sync.Mutex
	issues ValidationResult
	header string
}

func NewResultBuilder(header string) *IssueBuilder {
	return &IssueBuilder{
		issues: ValidationResult{},
		header: header,
	}
}

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
		issues.ErrorFormat = headeredErrListFormatFunc(fmt.Sprintf("- %s %q", bldr.header, field))
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
	for _, name := range issues.sortedKeys() {
		if issues[name] == nil {
			continue
		}
		msgs = append(msgs, issues[name].Error())
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

func (issues ValidationResult) sortedKeys() []string {
	keys := make([]string, 0, len(issues))
	for k := range issues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// headeredErrListFormatFunc is a basic formatter that outputs the errors as
// a bullet point list with a given header.
func headeredErrListFormatFunc(fieldName string) multierror.ErrorFormatFunc {
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
