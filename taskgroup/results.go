package taskgroup

import (
	"context"
	"fmt"
	"sync"
)

// Results collects successful return values from concurrent tasks.
//
// It behaves like Group: on the first error the shared context is
// canceled, panics are converted to errors, and Wait returns the first
// non-nil error along with the values gathered up to that point.
type Results[T any] struct {
	wg     sync.WaitGroup
	sem    chan struct{}
	cancel context.CancelCauseFunc
	ctx    context.Context

	mu     sync.Mutex
	values []T

	errOnce sync.Once
	err     error
}

// NewResults returns a Results group derived from ctx with no concurrency limit.
func NewResults[T any](ctx context.Context) *Results[T] {
	c, cancel := context.WithCancelCause(ctx)
	return &Results[T]{cancel: cancel, ctx: c}
}

// WithLimit caps concurrent in-flight tasks to n.
func (g *Results[T]) WithLimit(n int) *Results[T] {
	if n > 0 {
		g.sem = make(chan struct{}, n)
	}
	return g
}

// Context returns the group's context.
func (g *Results[T]) Context() context.Context { return g.ctx }

// Run launches fn. Successful results are appended (in completion order)
// to the slice returned by Wait.
func (g *Results[T]) Run(fn func(ctx context.Context) (T, error)) {
	if err := g.ctx.Err(); err != nil {
		return
	}
	if g.sem != nil {
		select {
		case g.sem <- struct{}{}:
		case <-g.ctx.Done():
			return
		}
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { <-g.sem }()
		}
		defer func() {
			if r := recover(); r != nil {
				g.recordErr(fmt.Errorf("taskgroup: panic: %v", r))
			}
		}()
		v, err := fn(g.ctx)
		if err != nil {
			g.recordErr(err)
			return
		}
		g.mu.Lock()
		g.values = append(g.values, v)
		g.mu.Unlock()
	}()
}

// Wait blocks until every Run task returns. The first non-nil error is
// reported alongside the values collected so far (in completion order).
// Wait cancels the group's context before returning so that long-lived
// consumers of Context() also exit.
func (g *Results[T]) Wait() ([]T, error) {
	g.wg.Wait()
	g.cancel(nil)
	return g.values, g.err
}

func (g *Results[T]) recordErr(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel(err)
	})
}

// Cause returns the first task error, when one was recorded.
func (g *Results[T]) Cause() error {
	return g.err
}
