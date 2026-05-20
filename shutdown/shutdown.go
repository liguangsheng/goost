// Package shutdown coordinates graceful shutdown hooks.
//
// Both an injectable Manager and a process-wide default are exposed.
// Most applications should use the package-level Add/Wait/Cleanup helpers
// for the default manager.
package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Manager collects shutdown hooks and runs them when triggered.
type Manager struct {
	mu       sync.Mutex
	hooks    []func()
	signals  []os.Signal
	logger   func(format string, args ...any)
	cleanup  sync.Once
	notifyCh chan os.Signal
}

// NewManager returns a Manager that listens for the given signals when
// Wait is called. If no signals are provided, SIGINT and SIGTERM are used.
func NewManager(signals ...os.Signal) *Manager {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	return &Manager{
		signals: signals,
		logger:  log.Printf,
	}
}

// SetLogger overrides the manager's logger. Defaults to log.Printf.
func (m *Manager) SetLogger(fn func(format string, args ...any)) {
	m.mu.Lock()
	if fn == nil {
		fn = func(string, ...any) {}
	}
	m.logger = fn
	m.mu.Unlock()
}

// Add registers fn to run during Cleanup. Hooks execute in registration order.
func (m *Manager) Add(fn func()) {
	m.mu.Lock()
	m.hooks = append(m.hooks, fn)
	m.mu.Unlock()
}

// Wait blocks until one of the configured signals arrives or ctx is canceled,
// then runs Cleanup. It returns the signal that fired, or nil on ctx done.
func (m *Manager) Wait(ctx context.Context) os.Signal {
	m.mu.Lock()
	if m.notifyCh == nil {
		m.notifyCh = make(chan os.Signal, 1)
		signal.Notify(m.notifyCh, m.signals...)
	}
	ch := m.notifyCh
	m.mu.Unlock()

	var sig os.Signal
	select {
	case sig = <-ch:
		m.logger("shutdown: received signal %s", sig)
	case <-ctx.Done():
		m.logger("shutdown: context canceled: %v", ctx.Err())
	}
	m.Cleanup()
	return sig
}

// Cleanup runs the registered hooks. Subsequent calls are no-ops. Panics
// from individual hooks are recovered and logged so later hooks still run.
func (m *Manager) Cleanup() {
	m.cleanup.Do(func() {
		m.mu.Lock()
		hooks := m.hooks
		logger := m.logger
		m.mu.Unlock()

		logger("shutdown: performing %d cleanups", len(hooks))
		for i, fn := range hooks {
			m.safeRun(i, fn, logger)
		}
		logger("shutdown: all cleanups done")
	})
}

func (m *Manager) safeRun(idx int, fn func(), logger func(string, ...any)) {
	defer func() {
		if r := recover(); r != nil {
			logger("shutdown: hook %d panicked: %v", idx, r)
		}
	}()
	fn()
}

var defaultManager = NewManager()

// Add appends fn to the default manager's hooks.
func Add(fn func()) { defaultManager.Add(fn) }

// Cleanup runs the default manager's hooks. Safe to call multiple times.
func Cleanup() { defaultManager.Cleanup() }

// Wait blocks on the default manager until a registered signal arrives,
// then runs Cleanup and returns. It does not exit the process.
func Wait(ctx context.Context) os.Signal { return defaultManager.Wait(ctx) }

// C blocks until a registered signal arrives, runs cleanups, and calls
// os.Exit(0). Provided for backwards compatibility; prefer Wait so the
// caller controls the exit code.
func C() {
	defaultManager.Wait(context.Background())
	os.Exit(0)
}
