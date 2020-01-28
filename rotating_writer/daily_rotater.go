package rotating_writer

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const oneday = time.Hour * 24

type DailyRotater struct {
	dir       string
	format    string
	maxBackup int

	file         *os.File
	rolloverAt   int64
	rolloverDate string

	m sync.Mutex
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

func (r *DailyRotater) ShouldRollover(current time.Time) bool {
	return current.Unix() > r.rolloverAt
}

func (r *DailyRotater) DoRollover(current time.Time) error {
	r.m.Lock()
	defer r.m.Unlock()

	if r.file != nil {
		r.file.Close()
	}

	file, err := r.open(r.newFilename(current))
	if err != nil {
		return err
	}

	r.file = file
	r.rolloverAt = r.nextRolloverAt(current)

	if r.maxBackup > 0 {
		r.deleteExpiredFiles()
	}
	return nil
}

func (r *DailyRotater) deleteExpiredFiles() {
	var toSort []os.FileInfo

	files, _ := ioutil.ReadDir(r.dir)
	for _, file := range files {
		// 跳过不符合命名规则的文件
		_, err := time.Parse(r.format, file.Name())
		if err != nil {
			continue
		}

		if file.Size() == 0 {
			// 删除空文件
			//os.Remove(filepath.Join(r.dir, file.Name()))
		} else {
			toSort = append(toSort, file)
		}
	}

	sort.Slice(toSort, func(i, j int) bool {
		a, _ := time.Parse(r.format, toSort[i].Name())
		b, _ := time.Parse(r.format, toSort[j].Name())
		return a.After(b)
	})

	for i, file := range toSort {
		if i >= r.maxBackup {
			os.Remove(filepath.Join(r.dir, file.Name()))
		}
	}
}

func (r *DailyRotater) newFilename(current time.Time) string {
	return filepath.Join(r.dir, current.Format(r.format))
}

func (r *DailyRotater) rolloverFilename(current time.Time) string {
	return filepath.Join(r.dir, current.Format(r.format))
}

func (r *DailyRotater) nextRolloverAt(current time.Time) int64 {
	return current.Add(oneday).Truncate(oneday).Unix()
}

func (r *DailyRotater) open(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666|os.ModeSticky)
}
