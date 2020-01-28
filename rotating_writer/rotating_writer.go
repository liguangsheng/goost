package rotating_writer

import (
	"io"
	"os"
	"time"
)

type Rotater interface {
	Writer() io.Writer
	ShouldRollover(time.Time) bool
	DoRollover(time.Time) error
}

type RotatingWriter struct {
	rotater Rotater
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	now := time.Now()
	if w.rotater.ShouldRollover(now) {
		if err := w.rotater.DoRollover(now); err != nil {
			return 0, err
		}
	}
	return w.rotater.Writer().Write(p)
}

func NewRotatingWriter(rotater Rotater) *RotatingWriter {
	return &RotatingWriter{rotater}
}

func NewDailyRotatingWriter(dir, format string, maxBackup int) *RotatingWriter {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	return NewRotatingWriter(NewDailyRotater(dir, format, maxBackup))
}
