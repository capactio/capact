package statusreporter

import (
	"context"

	"capact.io/capact/pkg/runner"
)

var _ runner.StatusReporter = &NoopReporter{}

// NoopReporter is a StatusReporter, which doesn't report anything anywhere.
// Used as a placeholder for non-built-in runners.
type NoopReporter struct {
}

// NewNoop returns new NoopReporter instance.
func NewNoop() *NoopReporter {
	return &NoopReporter{}
}

// Report does nothing and returns always nil.
func (n NoopReporter) Report(ctx context.Context, runnerCtx runner.Context, status interface{}) error {
	return nil
}
