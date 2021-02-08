package statusreporter

import (
	"context"

	"projectvoltron.dev/voltron/pkg/runner"
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

func (n NoopReporter) Report(ctx context.Context, runnerCtx runner.Context, status interface{}) error {
	return nil
}
