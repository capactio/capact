package printer

import "io"

var _ Status = &NoopStatusPrinter{}

// NoopStatusPrinter implements Status interface. It doesn't execute any operation on any method.
type NoopStatusPrinter struct{}

// NewNoopStatus returns a new instance of NoopStatusPrinter.
func NewNoopStatus() *NoopStatusPrinter {
	return &NoopStatusPrinter{}
}

// Step does nothing.
func (n *NoopStatusPrinter) Step(_ string, _ ...interface{}) {}

// End does nothing.
func (n *NoopStatusPrinter) End(_ bool) {}

// Infof does nothing.
func (n *NoopStatusPrinter) Infof(_ string, _ ...interface{}) {}

// InfoWithBody does nothing.
func (n *NoopStatusPrinter) InfoWithBody(_, _ string) {}

// Writer returns writer which discards all data.
func (n *NoopStatusPrinter) Writer() io.Writer {
	return io.Discard
}
