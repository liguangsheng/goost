// Package pool implements a bounded goroutine pool with optional queueing.
//
// The pool maintains up to size concurrent workers. A new worker is spawned
// on demand when all existing workers are busy and capacity remains. Tasks
// can either block, time out, or be queued, depending on the API used.
package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrScheduleTimeout is returned when ScheduleTimeout cannot schedule the
// task before its deadline.
var ErrScheduleTimeout = errors.New("schedule error: timed out")

// ErrPoolClosed is returned by Schedule and ScheduleTimeout after Close.
var ErrPoolClosed = errors.New("pool: closed")

// PanicHandler is called when a task panics. It runs on the worker's
// goroutine; do not block.
type PanicHandler func(panicVal any)

// Stats captures live pool metrics.
type Stats struct {
	Workers   int   // currently spawned workers
	InFlight  int   // tasks currently executing
	Queued    int   // tasks waiting in the buffer
	Completed int64 // tasks that finished (including panicked)
	Panics    int64 // tasks that panicked
}

// Pool is a bounded goroutine pool.
type Pool struct {
	sem     chan struct{}
	work    chan func()
	closed  chan struct{}
	mu      sync.RWMutex
	wg      sync.WaitGroup
	once    sync.Once
	onPanic atomic.Pointer[PanicHandler]

	workers   atomic.Int64
	inFlight  atomic.Int64
	completed atomic.Int64
	panics    atomic.Int64
}

// Option configures a Pool at construction time.
type Option func(*Pool)

// WithPanicHandler installs a callback to run after a task panics. If not
// set, panics are silently recovered so workers stay alive.
func WithPanicHandler(fn PanicHandler) Option {
	return func(p *Pool) {
		if fn == nil {
			return
		}
		p.onPanic.Store(&fn)
	}
}

// NewPool returns a pool with the given dimensions.
//   - size:  maximum concurrent workers.
//   - queue: queue depth for buffered tasks (0 = synchronous handoff).
//   - spawn: workers to start eagerly. Must be > 0 when queue > 0 so the
//     queue can be drained.
func NewPool(size, queue, spawn int, opts ...Option) (*Pool, error) {
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
	for _, opt := range opts {
		opt(p)
	}
	for range spawn {
		p.sem <- struct{}{}
		p.wg.Add(1)
		p.workers.Add(1)
		go p.worker(nil)
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

// ScheduleN schedules tasks sequentially and stops on the first
// scheduling error (e.g. pool closed). It returns the number of tasks
// that were accepted.
func (p *Pool) ScheduleN(tasks []func()) (n int, err error) {
	for _, t := range tasks {
		if err = p.Schedule(t); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (p *Pool) schedule(task func(), timeout <-chan time.Time) error {
	select {
	case <-p.closed:
		return ErrPoolClosed
	default:
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

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
		p.workers.Add(1)
		go p.worker(task)
		return nil
	}
}

// Close stops accepting new tasks and waits for in-flight workers to drain
// their queued work. Subsequent Schedule calls return ErrPoolClosed.
func (p *Pool) Close() {
	p.once.Do(func() {
		close(p.closed)
		p.mu.Lock()
		defer p.mu.Unlock()
		close(p.work)
	})
	p.wg.Wait()
}

// Stats returns a snapshot of pool metrics. Cheap to call.
func (p *Pool) Stats() Stats {
	return Stats{
		Workers:   int(p.workers.Load()),
		InFlight:  int(p.inFlight.Load()),
		Queued:    len(p.work),
		Completed: p.completed.Load(),
		Panics:    p.panics.Load(),
	}
}

func (p *Pool) worker(task func()) {
	defer func() {
		<-p.sem
		p.workers.Add(-1)
		p.wg.Done()
	}()

	if task != nil {
		p.safeRun(task)
	}
	for task := range p.work {
		p.safeRun(task)
	}
}

func (p *Pool) safeRun(task func()) {
	p.inFlight.Add(1)
	defer func() {
		p.inFlight.Add(-1)
		p.completed.Add(1)
		if r := recover(); r != nil {
			p.panics.Add(1)
			if h := p.onPanic.Load(); h != nil {
				(*h)(r)
			}
		}
	}()
	task()
}
