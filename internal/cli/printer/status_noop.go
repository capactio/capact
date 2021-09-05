package printer

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

// Info does nothing.
func (n *NoopStatusPrinter) Info(_, _ string) {}
