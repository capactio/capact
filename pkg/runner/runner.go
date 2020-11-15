package runner

// Runner provides generic runner service
type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func Execute(cfg Config) error {
	return nil
}
