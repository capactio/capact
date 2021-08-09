package validate

import (
	"fmt"
	"strings"
	"sync"
)

// ValidationResultAggregator allows you to aggregate validation result and generate single error.
// It's thread-safe.
type ValidationResultAggregator struct {
	issuesMsgs []string
	issuesCnt  int
	m          sync.RWMutex
}

// Report aggregates validation result. Can be used as wrapper:
//   err = rs.Report(validator.ValidateParameters(ctx, ifaceSchemas, params))
func (r *ValidationResultAggregator) Report(result ValidationResult, err error) error {
	r.m.Lock()
	defer r.m.Unlock()

	if err != nil {
		return err
	}
	r.issuesCnt += result.Len()
	if err := result.ErrorOrNil(); err != nil {
		r.issuesMsgs = append(r.issuesMsgs, err.Error())
	}
	return nil
}

// ErrorOrNil returns aggregated error for all reported issues. If no issues reported, returns nil.
func (r *ValidationResultAggregator) ErrorOrNil() error {
	if r == nil {
		return nil
	}
	r.m.RLock()
	defer r.m.RUnlock()

	switch r.issuesCnt {
	case 0:
	case 1:
		header := "1 validation error detected"
		return fmt.Errorf("%s:\n%s", header, strings.Join(r.issuesMsgs, "\n"))
	default:
		header := fmt.Sprintf("%d validation errors detected", r.issuesCnt)
		return fmt.Errorf("%s:\n%s", header, strings.Join(r.issuesMsgs, "\n"))
	}

	return nil
}
