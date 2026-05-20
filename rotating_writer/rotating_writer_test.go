package rotating_writer

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
		f.Close()
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
