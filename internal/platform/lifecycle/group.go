package lifecycle

import (
	"context"
	"errors"
	"sync"
)

type Logger interface {
	Printf(format string, v ...any)
}

type Runner struct {
	Name string
	Run  func(context.Context) error
}

type Group struct {
	logger  Logger
	runners []Runner
}

func NewGroup(logger Logger) *Group {
	return &Group{logger: logger}
}

func (g *Group) Add(name string, run func(context.Context) error) {
	g.runners = append(g.runners, Runner{Name: name, Run: run})
}

func (g *Group) Run(ctx context.Context) error {
	if len(g.runners) == 0 {
		<-ctx.Done()
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(g.runners))
	var wg sync.WaitGroup
	for _, runner := range g.runners {
		runner := runner
		if runner.Run == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.logf("runner %s started", runner.Name)
			err := runner.Run(runCtx)
			if err != nil && (errors.Is(err, context.Canceled) || runCtx.Err() != nil) {
				err = nil
			}
			if err != nil {
				g.logf("runner %s stopped with error: %v", runner.Name, err)
				cancel()
				errCh <- err
				return
			}
			g.logf("runner %s stopped", runner.Name)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		<-done
		return err
	case <-ctx.Done():
		cancel()
		<-done
		return nil
	case <-done:
		return nil
	}
}

func (g *Group) logf(format string, args ...any) {
	if g.logger != nil {
		g.logger.Printf(format, args...)
	}
}
