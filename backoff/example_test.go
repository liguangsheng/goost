package backoff_test

import (
	"fmt"
	"time"

	"github.com/liguangsheng/goost/backoff"
)

func ExampleBackoff_Next() {
	b := &backoff.Backoff{
		Initial: 100 * time.Millisecond,
		Max:     500 * time.Millisecond,
		Factor:  2,
	}

	fmt.Println(b.Next())
	fmt.Println(b.Next())
	fmt.Println(b.Next())
	fmt.Println(b.Next())

	// Output:
	// 100ms
	// 200ms
	// 400ms
	// 500ms
}
