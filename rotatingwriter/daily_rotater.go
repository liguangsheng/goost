package rotatingwriter

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const oneDay = 24 * time.Hour

// DailyRotater rolls a single file over to a new dated file once per day.
type DailyRotater struct {
	dir       string
	format    string
	maxBackup int

	file       *os.File
	rolloverAt int64
}

func NewDailyRotater(dir, format string, maxBackup int) *DailyRotater {
	return &DailyRotater{
		dir:       dir,
		format:    format,
		maxBackup: maxBackup,
	}
}

func (r *DailyRotater) Writer() io.Writer {
	return r.file
}

func (r *DailyRotater) ShouldRollover(current time.Time, _ int) bool {
	return r.file == nil || current.Unix() > r.rolloverAt
}

func (r *DailyRotater) DoRollover(current time.Time) error {
	file, err := r.open(r.filename(current))
	if err != nil {
		return err
	}

	if r.file != nil {
		_ = r.file.Close()
	}
	r.file = file
	r.rolloverAt = r.nextRolloverAt(current)

	if r.maxBackup > 0 {
		r.deleteExpiredFiles()
	}
	return nil
}

func (r *DailyRotater) deleteExpiredFiles() {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return
	}

	var matched []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, err := time.Parse(r.format, e.Name()); err != nil {
			continue
		}
		matched = append(matched, e)
	}

	sort.Slice(matched, func(i, j int) bool {
		a, _ := time.Parse(r.format, matched[i].Name())
		b, _ := time.Parse(r.format, matched[j].Name())
		return a.After(b)
	})

	for i, e := range matched {
		if i >= r.maxBackup {
			_ = os.Remove(filepath.Join(r.dir, e.Name()))
		}
	}
}

func (r *DailyRotater) filename(current time.Time) string {
	return filepath.Join(r.dir, current.Format(r.format))
}

func (r *DailyRotater) nextRolloverAt(current time.Time) int64 {
	return current.Add(oneDay).Truncate(oneDay).Unix()
}

func (r *DailyRotater) open(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
