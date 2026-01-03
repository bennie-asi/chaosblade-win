package exec

import "context"

// Runner defines execution behavior for experiments.
type Runner interface {
	Run(ctx context.Context) error
}

// NopRunner is a placeholder runner that performs no action.
type NopRunner struct{}

func (NopRunner) Run(ctx context.Context) error {
	return nil
}
