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

func (r *ValidationResultAggregator) ErrorOrNil() error {
	r.m.RLock()
	defer r.m.RUnlock()

	if len(r.issuesMsgs) > 0 {
		return fmt.Errorf("%d validation errors detected:\n%s", r.issuesCnt, strings.Join(r.issuesMsgs, "\n"))
	}
	return nil
}
