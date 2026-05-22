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
	maxAge    time.Duration

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

// WithMaxAge configures the rotater to delete dated files older than d
// (as encoded in the filename, not by mtime) at each rollover. Pass
// zero to disable age-based cleanup; this is the default.
//
// Combines additively with maxBackup: a file is deleted if EITHER
// limit is exceeded.
func (r *DailyRotater) WithMaxAge(d time.Duration) *DailyRotater {
	r.maxAge = d
	return r
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

	if r.maxBackup > 0 || r.maxAge > 0 {
		r.deleteExpiredFiles(current)
	}
	return nil
}

func (r *DailyRotater) deleteExpiredFiles(current time.Time) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return
	}

	type matchedEntry struct {
		path string
		date time.Time
	}
	var matched []matchedEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		d, err := time.Parse(r.format, e.Name())
		if err != nil {
			continue
		}
		matched = append(matched, matchedEntry{
			path: filepath.Join(r.dir, e.Name()),
			date: d,
		})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].date.After(matched[j].date)
	})

	currentPath := r.filename(current)
	backupIndex := 0
	for _, e := range matched {
		if e.path == currentPath {
			continue // never delete the file we just opened
		}
		drop := false
		if r.maxBackup > 0 && backupIndex >= r.maxBackup {
			drop = true
		}
		if r.maxAge > 0 && current.Sub(e.date) > r.maxAge {
			drop = true
		}
		if drop {
			_ = os.Remove(e.path)
		}
		backupIndex++
	}
}

func (r *DailyRotater) filename(current time.Time) string {
	return filepath.Join(r.dir, current.Format(r.format))
}

func (r *DailyRotater) nextRolloverAt(current time.Time) int64 {
	return current.Add(oneDay).Truncate(oneDay).Unix()
}

func (r *DailyRotater) open(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFilePerm) // #nosec G304 -- name is the caller-selected log path plus formatted date.
}
