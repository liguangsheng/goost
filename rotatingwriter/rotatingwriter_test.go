package rotatingwriter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DailyRotater_Write(t *testing.T) {
	dir := t.TempDir()
	w, err := NewDailyRotatingWriter(dir, "2006-01-02.log", 0)
	assert.NoError(t, err)

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
}

func Test_DailyRotater_MaxBackup(t *testing.T) {
	dir := t.TempDir()
	format := "2006-01-02.log"

	// Pre-create 5 dated files in the past.
	for i := 1; i <= 5; i++ {
		name := time.Now().AddDate(0, 0, -i).Format(format)
		f, err := os.Create(filepath.Join(dir, name))
		assert.NoError(t, err)
		_, _ = f.WriteString("x")
		_ = f.Close()
	}

	w, err := NewDailyRotatingWriter(dir, format, 2)
	assert.NoError(t, err)
	_, err = w.Write([]byte("hello"))
	assert.NoError(t, err)

	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	// Today's file plus 2 backups.
	assert.LessOrEqual(t, len(files), 3)
}

func Test_DailyRotater_MkdirAll(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	_, err := NewDailyRotatingWriter(dir, "2006-01-02.log", 0)
	assert.NoError(t, err)
}

func Test_SizeRotater_RotatesAtLimit(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "app.log")
	w, err := NewSizeRotatingWriter(base, 10, 3, false)
	assert.NoError(t, err)

	for range 5 {
		_, werr := w.Write([]byte("0123456789"))
		assert.NoError(t, werr)
	}

	files, err := filepath.Glob(filepath.Join(dir, "app.log*"))
	assert.NoError(t, err)
	// active + at most maxBackup
	assert.LessOrEqual(t, len(files), 4)
	assert.Contains(t, files, base)
}

func Test_DailyRotater_MaxAge(t *testing.T) {
	dir := t.TempDir()
	format := "2006-01-02.log"

	// Pre-create dated files: today-1, today-5, today-30.
	for _, days := range []int{1, 5, 30} {
		name := time.Now().AddDate(0, 0, -days).Format(format)
		f, err := os.Create(filepath.Join(dir, name))
		assert.NoError(t, err)
		_, _ = f.WriteString("x")
		_ = f.Close()
	}

	// MaxAge = 7 days: only today-1 and today-5 survive; today-30 is dropped.
	r := NewDailyRotater(dir, format, 0).WithMaxAge(7 * 24 * time.Hour)
	w := NewRotatingWriter(r)
	_, err := w.Write([]byte("hello"))
	assert.NoError(t, err)

	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.Name())
	}
	assert.Contains(t, names, time.Now().Format(format))
	assert.Contains(t, names, time.Now().AddDate(0, 0, -1).Format(format))
	assert.Contains(t, names, time.Now().AddDate(0, 0, -5).Format(format))
	assert.NotContains(t, names, time.Now().AddDate(0, 0, -30).Format(format))
}

func Test_DailyRotater_MaxAgeAndMaxBackup(t *testing.T) {
	dir := t.TempDir()
	format := "2006-01-02.log"

	// 6 files: today-1 .. today-6.
	for i := 1; i <= 6; i++ {
		name := time.Now().AddDate(0, 0, -i).Format(format)
		f, err := os.Create(filepath.Join(dir, name))
		assert.NoError(t, err)
		_ = f.Close()
	}

	// MaxBackup counts total files (active included): MaxBackup=4 keeps
	// 4 newest. MaxAge=4 days keeps anything dated > today-4 days.
	// Intersection: today + today-1..today-3 = 4 files.
	r := NewDailyRotater(dir, format, 4).WithMaxAge(4 * 24 * time.Hour)
	w := NewRotatingWriter(r)
	_, err := w.Write([]byte("x"))
	assert.NoError(t, err)

	files, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, files, 4)
}

func Test_SizeRotater_MaxAge(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "app.log")
	r, err := NewSizeRotater(base, 10, 0, false)
	assert.NoError(t, err)
	r.WithMaxAge(time.Hour)

	// Pre-create a stale backup at base.1 (mtime 2h ago) and a fresh
	// one at base.2 (current mtime). MaxAge=1h.
	stalePath := base + ".1"
	stale, _ := os.Create(stalePath)
	_ = stale.Close()
	past := time.Now().Add(-2 * time.Hour)
	assert.NoError(t, os.Chtimes(stalePath, past, past))

	freshPath := base + ".2"
	fresh, _ := os.Create(freshPath)
	_ = fresh.Close()

	// First write opens base; second write triggers shiftBackups.
	// shiftBackups:
	//   1. removeExpiredBackups removes the stale base.1
	//   2. shift loop renames remaining base.2 → base.3
	//   3. rename base → base.1
	// Final: base, base.1 (just rotated), base.3 (was the fresh one)
	w := NewRotatingWriter(r)
	_, _ = w.Write([]byte("0123456789"))
	_, _ = w.Write([]byte("0123456789"))

	files, _ := filepath.Glob(base + "*")
	assert.Len(t, files, 3, "stale removed, fresh shifted from .2 to .3, active becomes .1")
}

func Test_SizeRotater_Gzip(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "app.log")
	w, err := NewSizeRotatingWriter(base, 10, 2, true)
	assert.NoError(t, err)

	for range 4 {
		_, werr := w.Write([]byte("xxxxxxxxxx"))
		assert.NoError(t, werr)
	}

	files, err := filepath.Glob(filepath.Join(dir, "app.log.*.gz"))
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "expected at least one gzipped backup")
}
