package raindrop

import (
	"context"
	"sync"
)

type (
	Runner interface {
		Run(ctx context.Context)
	}

	RunnerGroup struct {
		sync.WaitGroup
	}
)

func (rg *RunnerGroup) Run(ctx context.Context, runner Runner) {
	rg.Add(1)
	go func() {
		defer rg.Done()
		runner.Run(ctx)
	}()
}
