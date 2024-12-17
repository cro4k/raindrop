package redistunnel

import "sync"

type (
	Runner interface {
		Run()
	}

	RunnerGroup struct {
		sync.WaitGroup
	}
)

func (rg *RunnerGroup) Run(runner Runner) {
	rg.Add(1)
	go func() {
		defer rg.Done()
		runner.Run()
	}()
}
