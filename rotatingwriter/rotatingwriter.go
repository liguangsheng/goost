// Package rotatingwriter provides an io.Writer that rotates its backing
// file according to a Rotater strategy (e.g. daily, size-based).
package rotatingwriter

import (
	"io"
	"os"
	"sync"
	"time"
)

// Rotater decides when to rotate and exposes the current writer.
type Rotater interface {
	Writer() io.Writer
	ShouldRollover(time.Time, int) bool // n = bytes about to be written
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
	if w.rotater.ShouldRollover(now, len(p)) {
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

// NewSizeRotatingWriter is a convenience constructor for size-based rotation.
// Each rollover creates a new file named base.N (N increments) up to
// maxBackup, oldest deleted. If gzip is true, rolled files are gzipped.
func NewSizeRotatingWriter(base string, maxBytes int64, maxBackup int, gzip bool) (*RotatingWriter, error) {
	r, err := NewSizeRotater(base, maxBytes, maxBackup, gzip)
	if err != nil {
		return nil, err
	}
	return NewRotatingWriter(r), nil
}
