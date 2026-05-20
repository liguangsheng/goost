package rotating_writer

import (
	"io"
	"os"
	"sync"
	"time"
)

// Rotater decides when to rotate and exposes the current writer.
type Rotater interface {
	Writer() io.Writer
	ShouldRollover(time.Time) bool
	DoRollover(time.Time) error
}

// RotatingWriter wraps a Rotater and serializes Write calls so the rollover
// check and the write itself are atomic.
type RotatingWriter struct {
	rotater Rotater
	mu      sync.Mutex
}

func NewRotatingWriter(rotater Rotater) *RotatingWriter {
	return &RotatingWriter{rotater: rotater}
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	if w.rotater.ShouldRollover(now) {
		if err := w.rotater.DoRollover(now); err != nil {
			return 0, err
		}
	}
	return w.rotater.Writer().Write(p)
}

// NewDailyRotatingWriter is a convenience constructor for the common
// "one file per day" case. dir is created if missing.
func NewDailyRotatingWriter(dir, format string, maxBackup int) (*RotatingWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return NewRotatingWriter(NewDailyRotater(dir, format, maxBackup)), nil
}
