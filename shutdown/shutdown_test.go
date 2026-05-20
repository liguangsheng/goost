package shutdown

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_CleanupRunsHooks(t *testing.T) {
	m := NewManager()
	m.SetLogger(nil)

	var a, b atomic.Int64
	m.Add(func() { a.Add(1) })
	m.Add(func() { b.Add(1) })
	m.Cleanup()
	assert.EqualValues(t, 1, a.Load())
	assert.EqualValues(t, 1, b.Load())

	// Idempotent.
	m.Cleanup()
	assert.EqualValues(t, 1, a.Load())
}

func Test_CleanupRecoversPanic(t *testing.T) {
	m := NewManager()
	m.SetLogger(nil)

	var after atomic.Int64
	m.Add(func() { panic("boom") })
	m.Add(func() { after.Add(1) })
	assert.NotPanics(t, func() { m.Cleanup() })
	assert.EqualValues(t, 1, after.Load())
}

func Test_WaitOnSignal(t *testing.T) {
	m := NewManager(syscall.SIGUSR1)
	m.SetLogger(nil)

	var ran atomic.Int64
	m.Add(func() { ran.Add(1) })

	done := make(chan struct{})
	go func() {
		m.Wait(context.Background())
		close(done)
	}()

	// Give Wait time to install the notify channel.
	time.Sleep(20 * time.Millisecond)
	assert.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGUSR1))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Wait did not return on signal")
	}
	assert.EqualValues(t, 1, ran.Load())
}

func Test_HookTimeout(t *testing.T) {
	m := NewManager()
	m.SetLogger(nil)

	var finished atomic.Int64
	m.Add(func() {
		time.Sleep(100 * time.Millisecond)
		finished.Add(1)
	}, WithTimeout(20*time.Millisecond), WithName("slow"))

	var after atomic.Int64
	m.Add(func() { after.Add(1) })

	start := time.Now()
	m.Cleanup()
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 80*time.Millisecond, "Cleanup should abandon slow hook")
	assert.EqualValues(t, 1, after.Load(), "subsequent hooks still run")
}

func Test_WaitOnContextCancel(t *testing.T) {
	m := NewManager(syscall.SIGUSR2)
	m.SetLogger(nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sig := m.Wait(ctx)
	assert.Nil(t, sig)
}
