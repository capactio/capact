package validation

import (
	"fmt"
	"strings"
	"sync"
)

// ResultAggregator allows you to aggregate validation result and generate single error.
// It's thread-safe.
type ResultAggregator struct {
	issuesMsgs []string
	issuesCnt  int
	m          sync.RWMutex
}

// Report aggregates validation result. Can be used as wrapper:
//   err = rs.Report(validator.ValidateParameters(ctx, ifaceSchemas, params))
func (r *ResultAggregator) Report(result Result, err error) error {
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
func (r *ResultAggregator) ErrorOrNil() error {
	if r == nil {
		return nil
	}
	r.m.RLock()
	defer r.m.RUnlock()

	if len(r.issuesMsgs) > 0 {
		return fmt.Errorf("%d validation %s detected:\n%s",
			r.issuesCnt, nounFor("error", r.issuesCnt), strings.Join(r.issuesMsgs, "\n"))
	}
	return nil
}

func nounFor(str string, numberOfItems int) string {
	if numberOfItems == 1 {
		return str
	}

	return str + "s"
}
