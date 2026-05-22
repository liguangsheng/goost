// Package taskgroup is a small alternative to golang.org/x/sync/errgroup
// that adds a concurrency limit and panic recovery.
//
// On the first non-nil error returned from a Go-spawned task, the group's
// context is canceled. Wait returns that first error. Panics in tasks are
// converted to errors so a single panic does not crash the program.
package taskgroup

import (
	"context"
	"fmt"
	"sync"
)

// Group is a collection of goroutines working on subtasks that are part
// of the same overall task. The zero value is invalid; use New.
type Group struct {
	wg     sync.WaitGroup
	sem    chan struct{}
	cancel context.CancelCauseFunc
	ctx    context.Context

	errOnce sync.Once
	err     error
}

// New returns a Group derived from ctx with no concurrency limit.
// Cancel ctx (or any task returning an error) cancels every task.
func New(ctx context.Context) *Group {
	c, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel, ctx: c}
}

// WithLimit caps the number of tasks executing concurrently to n.
// Subsequent Go calls block until a slot is free.
func (g *Group) WithLimit(n int) *Group {
	if n > 0 {
		g.sem = make(chan struct{}, n)
	}
	return g
}

// Context returns the group's context. Cancelled on the first task error
// or when Wait returns.
func (g *Group) Context() context.Context { return g.ctx }

// Go runs fn in its own goroutine, respecting the concurrency limit.
// Returns immediately if the group is already cancelled.
func (g *Group) Go(fn func(ctx context.Context) error) {
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
		if err := fn(g.ctx); err != nil {
			g.recordErr(err)
		}
	}()
}

// Wait blocks until every Go-spawned task returns and reports the first
// non-nil error (if any). Wait cancels the group's context before
// returning so that long-lived consumers of Context() also exit.
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel(nil)
	return g.err
}

func (g *Group) recordErr(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel(err)
	})
}

// Cause returns the cancellation cause of g.Context(). Useful when chained
// downstream contexts want to surface the original failing task's error.
func (g *Group) Cause() error {
	return g.err
}
