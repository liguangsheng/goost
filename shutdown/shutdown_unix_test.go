//go:build !windows

package shutdown

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_WaitOnSignal(t *testing.T) {
	m := NewManager(syscall.SIGUSR1)
	m.SetLogger(nil)

	var ran atomic.Int64
	m.Add(func() { ran.Add(1) })

	done := make(chan os.Signal, 1)
	go func() {
		done <- m.Wait(context.Background())
	}()

	// Give Wait time to install the notify channel.
	time.Sleep(20 * time.Millisecond)
	assert.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGUSR1))

	select {
	case sig := <-done:
		assert.Equal(t, syscall.SIGUSR1, sig)
	case <-time.After(time.Second):
		t.Fatal("Wait did not return on signal")
	}
	assert.EqualValues(t, 1, ran.Load())
}
