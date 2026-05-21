package keyedmutex_test

import (
	"context"
	"fmt"

	"github.com/liguangsheng/goost/keyedmutex"
)

func ExampleMutex_WithLock() {
	m := keyedmutex.New[string]()
	_ = m.WithLock(context.Background(), "user:42", func() error {
		fmt.Println("recompute user:42")
		return nil
	})

	fmt.Println("active keys:", m.Len())

	// Output:
	// recompute user:42
	// active keys: 0
}
