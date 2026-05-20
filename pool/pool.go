// Package pool implements a bounded goroutine pool with optional queueing.
//
// The pool maintains up to size concurrent workers. A new worker is spawned
// on demand when all existing workers are busy and capacity remains. Tasks
// can either block, time out, or be queued, depending on the API used.
package pool

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrScheduleTimeout is returned when ScheduleTimeout cannot schedule the
// task before its deadline.
var ErrScheduleTimeout = errors.New("schedule error: timed out")

// ErrPoolClosed is returned by Schedule and ScheduleTimeout after Close.
var ErrPoolClosed = errors.New("pool: closed")

// Pool is a bounded goroutine pool.
type Pool struct {
	sem    chan struct{}
	work   chan func()
	closed chan struct{}
	wg     sync.WaitGroup
	once   sync.Once
}

// NewPool returns a pool with the given dimensions.
//   - size:  maximum concurrent workers.
//   - queue: queue depth for buffered tasks (0 = synchronous handoff).
//   - spawn: workers to start eagerly. Must be > 0 when queue > 0 so the
//     queue can be drained.
func NewPool(size, queue, spawn int) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("pool: size must be > 0")
	}
	if spawn < 0 {
		return nil, errors.New("pool: spawn must be >= 0")
	}
	if spawn == 0 && queue > 0 {
		return nil, errors.New("pool: spawn must be > 0 when queue > 0")
	}
	if spawn > size {
		return nil, errors.New("pool: spawn must be <= size")
	}
	p := &Pool{
		sem:    make(chan struct{}, size),
		work:   make(chan func(), queue),
		closed: make(chan struct{}),
	}
	for i := 0; i < spawn; i++ {
		p.sem <- struct{}{}
		p.wg.Add(1)
		go p.worker(func() {})
	}
	return p, nil
}

// Schedule runs task on the pool. It blocks until a worker is available or
// the pool is closed.
func (p *Pool) Schedule(task func()) error {
	return p.schedule(task, nil)
}

// ScheduleTimeout is like Schedule but returns ErrScheduleTimeout when the
// task cannot be scheduled before timeout elapses.
func (p *Pool) ScheduleTimeout(timeout time.Duration, task func()) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	return p.schedule(task, timer.C)
}

func (p *Pool) schedule(task func(), timeout <-chan time.Time) error {
	select {
	case <-p.closed:
		return ErrPoolClosed
	default:
	}

	select {
	case <-p.closed:
		return ErrPoolClosed
	case <-timeout:
		return ErrScheduleTimeout
	case p.work <- task:
		return nil
	case p.sem <- struct{}{}:
		p.wg.Add(1)
		go p.worker(task)
		return nil
	}
}

// Close stops accepting new tasks and waits for in-flight workers to drain
// their queued work. Subsequent Schedule calls return ErrPoolClosed.
func (p *Pool) Close() {
	p.once.Do(func() {
		close(p.closed)
		close(p.work)
	})
	p.wg.Wait()
}

func (p *Pool) worker(task func()) {
	defer func() {
		<-p.sem
		p.wg.Done()
	}()

	safeRun(task)
	for task := range p.work {
		safeRun(task)
	}
}

func safeRun(task func()) {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("pool: worker recovered from panic", zap.Any("panic", r))
		}
	}()
	task()
}
