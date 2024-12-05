package pool

import (
	"fmt"
	"time"
)

var ErrScheduleTimeout = fmt.Errorf("schedule error: timed out")

type Pool struct {
	sem  chan struct{}
	work chan func()
}

func NewPool(size, queue, spawn int) (*Pool, error) {
	if spawn <= 0 && queue > 0 {
		return nil, fmt.Errorf("spawn must be greater than 0 when queue is set")
	}
	if spawn > size {
		return nil, fmt.Errorf("spawn must be less or equal to size")
	}
	p := &Pool{
		sem:  make(chan struct{}, size),
		work: make(chan func(), queue),
	}
	for i := 0; i < spawn; i++ {
		p.sem <- struct{}{}
		go p.worker(func() {})
	}

	return p, nil
}

func (p *Pool) Schedule(task func()) {
	p.schedule(task, nil)
}

func (p *Pool) ScheduleTimeout(timeout time.Duration, task func()) error {
	return p.schedule(task, time.After(timeout))
}

func (p *Pool) schedule(task func(), timeout <-chan time.Time) error {
	select {
	case <-timeout:
		return ErrScheduleTimeout
	case p.work <- task:
		return nil
	case p.sem <- struct{}{}:
		go p.worker(task)
		return nil
	}
}

func (p *Pool) worker(task func()) {
	defer func() { <-p.sem }()

	task()

	for task := range p.work {
		task()
	}
}
