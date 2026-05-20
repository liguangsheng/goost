package rotatingwriter

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SizeRotater rolls over when the current file's size plus the pending
// write would exceed maxBytes. The active file is base; rotated files are
// numbered base.1, base.2, ... up to maxBackup, with base.1 always being
// the most recent.
type SizeRotater struct {
	base      string
	maxBytes  int64
	maxBackup int
	gzipBak   bool

	file *os.File
	size int64
}

// NewSizeRotater returns a SizeRotater. maxBytes must be > 0.
func NewSizeRotater(base string, maxBytes int64, maxBackup int, gzipBak bool) (*SizeRotater, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("rotatingwriter: maxBytes must be > 0")
	}
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		return nil, err
	}
	return &SizeRotater{
		base:      base,
		maxBytes:  maxBytes,
		maxBackup: maxBackup,
		gzipBak:   gzipBak,
	}, nil
}

func (r *SizeRotater) Writer() io.Writer { return sizeWriter{r} }

func (r *SizeRotater) ShouldRollover(_ time.Time, n int) bool {
	return r.file == nil || r.size+int64(n) > r.maxBytes
}

func (r *SizeRotater) DoRollover(_ time.Time) error {
	if r.file != nil {
		_ = r.file.Close()
		if err := r.shiftBackups(); err != nil {
			return err
		}
		r.file = nil
	}
	f, err := os.OpenFile(r.base, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	r.file = f
	r.size = 0
	return nil
}

func (r *SizeRotater) shiftBackups() error {
	ext := ""
	if r.gzipBak {
		ext = ".gz"
	}

	// Remove oldest if at limit.
	if r.maxBackup > 0 {
		oldest := fmt.Sprintf("%s.%d%s", r.base, r.maxBackup, ext)
		_ = os.Remove(oldest)
	}

	// Shift base.N → base.(N+1).
	type pair struct {
		idx  int
		path string
	}
	var existing []pair
	dir := filepath.Dir(r.base)
	prefix := filepath.Base(r.base) + "."
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		idxStr := strings.TrimPrefix(name, prefix)
		idxStr = strings.TrimSuffix(idxStr, ext)
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			continue
		}
		existing = append(existing, pair{idx, filepath.Join(dir, name)})
	}
	sort.Slice(existing, func(i, j int) bool { return existing[i].idx > existing[j].idx })
	for _, p := range existing {
		if r.maxBackup > 0 && p.idx >= r.maxBackup {
			_ = os.Remove(p.path)
			continue
		}
		next := fmt.Sprintf("%s.%d%s", r.base, p.idx+1, ext)
		_ = os.Rename(p.path, next)
	}

	// Move base → base.1 (and gzip if requested).
	target1 := fmt.Sprintf("%s.1%s", r.base, ext)
	if r.gzipBak {
		if err := gzipFile(r.base, target1); err != nil {
			return err
		}
		_ = os.Remove(r.base)
	} else {
		if err := os.Rename(r.base, target1); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func gzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	gw := gzip.NewWriter(out)
	if _, err := io.Copy(gw, in); err != nil {
		return err
	}
	return gw.Close()
}

// sizeWriter wraps the rotater so writes update size accounting.
type sizeWriter struct{ r *SizeRotater }

func (w sizeWriter) Write(p []byte) (int, error) {
	n, err := w.r.file.Write(p)
	w.r.size += int64(n)
	return n, err
}
