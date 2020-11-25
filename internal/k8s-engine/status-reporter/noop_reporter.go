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

// NewK8sConfigMap returns new K8sConfigMapReporter instance.
func NewNoop() *NoopReporter {
	return &NoopReporter{}
}

func (n NoopReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	return nil
}
