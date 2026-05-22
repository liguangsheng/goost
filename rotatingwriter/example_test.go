package rotatingwriter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/liguangsheng/goost/rotatingwriter"
)

func ExampleNewSizeRotatingWriter() {
	dir, err := os.MkdirTemp("", "rotatingwriter-example-*")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	base := filepath.Join(dir, "app.log")
	w, err := rotatingwriter.NewSizeRotatingWriter(base, 5, 2, false)
	if err != nil {
		panic(err)
	}

	for range 3 {
		if _, writeErr := w.Write([]byte("hello")); writeErr != nil {
			panic(writeErr)
		}
	}

	files, err := filepath.Glob(filepath.Join(dir, "app.log*"))
	if err != nil {
		panic(err)
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, filepath.Base(file))
	}
	slices.Sort(names)
	fmt.Println(names)

	// Output:
	// [app.log app.log.1 app.log.2]
}
